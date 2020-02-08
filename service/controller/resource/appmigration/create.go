package appmigration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/cluster-operator/pkg/annotation"
	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	awskey "github.com/giantswarm/cluster-operator/service/controller/aws/key"
	azurekey "github.com/giantswarm/cluster-operator/service/controller/azure/key"
	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
	pkgerrors "github.com/giantswarm/cluster-operator/service/controller/internal/errors"
	"github.com/giantswarm/cluster-operator/service/controller/key"
	kvmkey "github.com/giantswarm/cluster-operator/service/controller/kvm/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	objectMeta, err := r.getClusterObjectMetaFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// Cluster is being deleted. No migration is necessary.
	if key.IsDeleted(objectMeta) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "cluster is being deleted")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	chartSpecsToMigrate := r.newChartSpecsToMigrate()

	if len(chartSpecsToMigrate) == 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no charts to migrate")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	cr, err := r.getClusterConfigFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	if cc.Client.TenantCluster.G8s == nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant clients not available")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", label.ManagedBy, project.Name()),
	}

	ch := make(chan response, 1)

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	go func() {
		chartConfigs, err := cc.Client.TenantCluster.G8s.CoreV1alpha1().ChartConfigs("giantswarm").List(listOptions)
		ch <- response{
			ChartConfigs: chartConfigs.Items,
			Error:        err,
		}
	}()

	select {
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			// Set status so we don't try to connect to the tenant cluster
			// again in this reconciliation loop.
			cc.Status.TenantCluster.IsUnavailable = true

			r.logger.LogCtx(ctx, "level", "debug", "message", "timeout getting chartconfig crs")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}
	default:
		// Fall through.
	}

	res := <-ch

	if tenant.IsAPINotAvailable(res.Error) {
		// Set status so we don't try to connect to the tenant cluster
		// again in this reconciliation loop.
		cc.Status.TenantCluster.IsUnavailable = true

		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is not available")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	} else if pkgerrors.IsChartConfigNotAvailable(res.Error) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "chartconfig CRs are not available")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	} else if pkgerrors.IsChartConfigNotInstalled(res.Error) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "chartconfig CRD does not exist")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	} else if res.Error != nil {
		return microerror.Mask(err)
	}

	for _, chartSpec := range chartSpecsToMigrate {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding out if chartconfig CR %#q has been migrated", chartSpec.ChartName))

		chartCR, err := getChartConfigByName(res.ChartConfigs, chartSpec.ChartName)
		if IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("chartconfig CR %#q has been migrated, continuing", chartSpec.ChartName))
			continue
		} else if err != nil {
			return microerror.Mask(err)
		}

		// Cordon chartconfig CR so no changes are applied.
		_, ok := chartCR.Annotations[annotation.CordonReason]
		if !ok {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("cordoning chartconfig CR %#q", chartSpec.ChartName))

			err = patchChartConfig(cc.Client.TenantCluster.G8s, chartCR, addCordonAnnotations())
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("cordoned chartconfig CR %#q", chartSpec.ChartName))
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("chartconfig CR %#q is already cordoned", chartSpec.ChartName))
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding out if app CR %#q is deployed", chartSpec.AppName))

		appCR, err := r.g8sClient.ApplicationV1alpha1().Apps(key.ClusterID(cr)).Get(chartSpec.AppName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("app CR %#q does not exist yet, continuing", chartSpec.AppName))
			continue
		}

		if appCR.Status.Release.Status == "DEPLOYED" {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("app CR %#q has status %#q", chartSpec.AppName, appCR.Status.Release.Status))
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("adding annotation for deleting chartconfig CR %#q", chartSpec.ChartName))

			// Add deletion annotation which will trigger chart-operator to
			// delete the chartconfig CR but not the Helm release.
			err = patchChartConfig(cc.Client.TenantCluster.G8s, chartCR, addDeleteAnnotation())
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("added annotation to chartconfig CR %#q", chartSpec.ChartName))
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting chartconfig CR %#q", chartSpec.ChartName))

			// Lastly delete the chartconfig CR.
			err = cc.Client.TenantCluster.G8s.CoreV1alpha1().ChartConfigs("giantswarm").Delete(chartCR.Name, &metav1.DeleteOptions{})
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted chartconfig CR %#q", chartSpec.ChartName))
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("app CR %#q has status %#q, continuing", chartSpec.AppName, appCR.Status.Release.Status))
			continue
		}
	}

	return nil
}

func (r *Resource) newChartSpecsToMigrate() []key.ChartSpec {
	chartSpecs := []key.ChartSpec{}

	switch r.provider {
	case "aws":
		chartSpecs = append(key.CommonChartSpecs(), awskey.ChartSpecs()...)
	case "azure":
		chartSpecs = append(key.CommonChartSpecs(), azurekey.ChartSpecs()...)
	case "kvm":
		chartSpecs = append(key.CommonChartSpecs(), kvmkey.ChartSpecs()...)
	default:
		chartSpecs = key.CommonChartSpecs()
	}

	chartSpecsToMigrate := []key.ChartSpec{}

	for _, spec := range chartSpecs {
		chartSpecsToMigrate = append(chartSpecsToMigrate, spec)
	}

	return chartSpecsToMigrate
}

func addCordonAnnotations() map[string]string {
	return map[string]string{
		annotation.CordonReason:    "cordoning chartconfig CR for migration to app CR",
		annotation.CordonUntilDate: key.CordonUntilDate(),
	}
}

func addDeleteAnnotation() map[string]string {
	return map[string]string{
		annotation.DeleteCustomResourceOnly: "true",
	}
}

func patchChartConfig(tenantG8sClient versioned.Interface, chartCR v1alpha1.ChartConfig, annotations map[string]string) error {
	patches := []Patch{}

	if len(chartCR.Annotations) == 0 {
		patches = append(patches, Patch{
			Op:    "add",
			Path:  "/metadata/annotations",
			Value: map[string]string{},
		})
	}

	for k, v := range annotations {
		patches = append(patches, Patch{
			Op:    "add",
			Path:  fmt.Sprintf("/metadata/annotations/%s", replaceToEscape(k)),
			Value: v,
		})
	}

	bytes, err := json.Marshal(patches)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = tenantG8sClient.CoreV1alpha1().ChartConfigs("giantswarm").Patch(chartCR.Name, types.JSONPatchType, bytes)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func replaceToEscape(from string) string {
	return strings.Replace(from, "/", "~1", -1)
}
