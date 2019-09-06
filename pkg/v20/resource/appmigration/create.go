package appmigration

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/cluster-operator/pkg/annotation"
	"github.com/giantswarm/cluster-operator/pkg/v20/key"
	awskey "github.com/giantswarm/cluster-operator/service/controller/aws/v20/key"
	azurekey "github.com/giantswarm/cluster-operator/service/controller/azure/v20/key"
	kvmkey "github.com/giantswarm/cluster-operator/service/controller/kvm/v20/key"
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
		resourcecanceledcontext.SetCanceled(ctx)

		return nil
	}

	clusterConfig, err := r.getClusterConfigFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	tenantAPIDomain, err := key.APIDomain(clusterConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	tenantG8sClient, err := r.tenant.NewG8sClient(ctx, clusterConfig.ID, tenantAPIDomain)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, chartSpec := range r.newChartSpecs() {
		// Only migrate chartconfigs that have app CRs.
		if chartSpec.HasAppCR == true {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding out if chartconfig CR %#q has been migrated", chartSpec.ChartName))

			chartCR, err := tenantG8sClient.CoreV1alpha1().ChartConfigs("giantswarm").Get(chartSpec.ChartName, metav1.GetOptions{})
			if tenant.IsAPINotAvailable(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is not available yet")
				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
				resourcecanceledcontext.SetCanceled(ctx)
				return nil
			} else if apierrors.IsNotFound(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("chartconfig CR %#q has been migrated, continuing", chartSpec.ChartName))
			} else if err != nil {
				return microerror.Mask(err)
			}

			// Cordon chartconfig CR so no changes are applied.
			_, ok := chartCR.Annotations[annotation.CordonReason]
			if !ok {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("cordoning chartconfig CR %#q", chartSpec.ChartName))

				err = patchChartConfig(tenantG8sClient, chartCR, addCordonAnnotations())
				if err != nil {
					return microerror.Mask(err)
				}

				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("cordoned chartconfig CR %#q", chartSpec.ChartName))
			} else {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("chartconfig CR %#q is already cordoned", chartSpec.ChartName))
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding out app CR %#q is deployed", chartSpec.AppName))

			// Check if there is a deployed app CR.
			appCR, err := r.g8sClient.ApplicationV1alpha1().Apps(clusterConfig.ID).Get(chartSpec.AppName, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("app CR %#q does not exist yet, continuing", chartSpec.AppName))
				continue
			} else if err != nil {
				return microerror.Mask(err)
			}

			if appCR.Status.Release != nil && appCR.Status.Release.Status == "DEPLOYED" {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("app CR %#q has status %#q", chartSpec.AppName, appCR.Status.Release.Status))
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("adding annotation for deleting chartconfig CR %#q", chartSpec.ChartName))

				// Add deletion annotation which will trigger chart-operator to
				// delete the chartconfig CR but not the Helm release.
				err = patchChartConfig(tenantG8sClient, chartCR, addDeleteAnnotation())
				if err != nil {
					return microerror.Mask(err)
				}

				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("added annotation to chartconfig CR %#q", chartSpec.ChartName))
			} else {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("app CR %#q has status %#q, continuing", chartSpec.AppName, appCR.Status.Release.Status))
				continue
			}
		}
	}

	return nil
}

func (r *Resource) newChartSpecs() []key.ChartSpec {
	switch r.provider {
	case "aws":
		return append(key.CommonChartSpecs(), awskey.ChartSpecs()...)
	case "azure":
		return append(key.CommonChartSpecs(), azurekey.ChartSpecs()...)
	case "kvm":
		return append(key.CommonChartSpecs(), kvmkey.ChartSpecs()...)
	default:
		return key.CommonChartSpecs()
	}
}

func addCordonAnnotations() map[string]string {
	return map[string]string{
		annotation.CordonReason:    "cordoning chartconfig CR for migration to app CR",
		annotation.CordonUntilDate: time.Now().Add(1 * time.Hour).Format("2006-01-02T15:04:05"),
	}
}

func addDeleteAnnotation() map[string]string {
	return map[string]string{
		annotation.DeleteCustomResourceOnly: "true",
	}
}

func patchChartConfig(tenantG8sClient versioned.Interface, chartCR *v1alpha1.ChartConfig, annotations map[string]string) error {
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
