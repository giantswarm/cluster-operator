package appmigration

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
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

	chartSpecsToMigrate := r.newChartSpecsToMigrate()

	if len(chartSpecsToMigrate) == 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no charts to migrate")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil
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
		if spec.HasAppCR {
			chartSpecsToMigrate = append(chartSpecsToMigrate, spec)
		}
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
