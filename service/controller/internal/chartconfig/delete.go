package chartconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/resource/crud"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
)

func (c *ChartConfig) ApplyDeleteChange(ctx context.Context, clusterConfig ClusterConfig, chartConfigsToDelete []*v1alpha1.ChartConfig) error {
	if len(chartConfigsToDelete) > 0 {
		cc, err := controllercontext.FromContext(ctx)
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", "deleting chartconfigs")

		for _, chartConfig := range chartConfigsToDelete {
			err := cc.Client.TenantCluster.G8s.CoreV1alpha1().ChartConfigs(resourceNamespace).Delete(chartConfig.Name, &metav1.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", "deleted chartconfigs")
	} else {
		c.logger.LogCtx(ctx, "level", "debug", "message", "no need to delete chartconfigs")
	}

	return nil

}

// NewDeletePatch is a no-op because chartconfig CRs in the tenant cluster are
// deleted with the tenant cluster resources.
func (c *ChartConfig) NewDeletePatch(ctx context.Context, currentState, desiredState []*v1alpha1.ChartConfig) (*crud.Patch, error) {
	return nil, nil
}

func (c *ChartConfig) newDeleteChangeForUpdatePatch(ctx context.Context, currentChartConfigs, desiredChartConfigs []*v1alpha1.ChartConfig) ([]*v1alpha1.ChartConfig, error) {
	c.logger.LogCtx(ctx, "level", "debug", "message", "finding out which chartconfigs have to be deleted")

	chartConfigsToDelete := make([]*v1alpha1.ChartConfig, 0)

	for _, currentChartConfig := range currentChartConfigs {
		_, err := getChartConfigByName(desiredChartConfigs, currentChartConfig.Name)
		// Existing ChartConfig is not desired anymore so it should be deleted.
		if IsNotFound(err) {
			chartConfigsToDelete = append(chartConfigsToDelete, currentChartConfig)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d chartconfigs that have to be deleted", len(chartConfigsToDelete)))

	return chartConfigsToDelete, nil
}
