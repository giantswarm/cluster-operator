package chartconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"

	"github.com/giantswarm/cluster-operator/pkg/v21/controllercontext"
)

func (c *ChartConfig) ApplyUpdateChange(ctx context.Context, clusterConfig ClusterConfig, chartConfigsToUpdate []*v1alpha1.ChartConfig) error {
	if len(chartConfigsToUpdate) > 0 {
		cc, err := controllercontext.FromContext(ctx)
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", "updating chartconfigs")

		for _, chartConfigToUpdate := range chartConfigsToUpdate {
			_, err := cc.Client.TenantCluster.G8s.CoreV1alpha1().ChartConfigs(resourceNamespace).Update(chartConfigToUpdate)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", "updated chartconfigs")
	} else {
		c.logger.LogCtx(ctx, "level", "debug", "message", "no need to update chartconfigs")
	}

	return nil
}

func (c *ChartConfig) NewUpdatePatch(ctx context.Context, currentState, desiredState []*v1alpha1.ChartConfig) (*controller.Patch, error) {
	create, err := c.newCreateChange(ctx, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	update, err := c.newUpdateChange(ctx, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	delete, err := c.newDeleteChangeForUpdatePatch(ctx, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (c *ChartConfig) newUpdateChange(ctx context.Context, currentChartConfigs, desiredChartConfigs []*v1alpha1.ChartConfig) ([]*v1alpha1.ChartConfig, error) {
	c.logger.LogCtx(ctx, "level", "debug", "message", "finding out which chartconfigs have to be updated")

	chartConfigsToUpdate := make([]*v1alpha1.ChartConfig, 0)

	for _, currentChartConfig := range currentChartConfigs {
		desiredChartConfig, err := getChartConfigByName(desiredChartConfigs, currentChartConfig.Name)
		if IsNotFound(err) {
			// Ignore here. These are handled by newDeleteChangeForUpdatePatch().
			continue
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		if isChartConfigModified(desiredChartConfig, currentChartConfig) {
			// Make a copy and set the resource version so the CR can be updated.
			chartConfigToUpdate := desiredChartConfig.DeepCopy()
			chartConfigToUpdate.ObjectMeta.ResourceVersion = currentChartConfig.ObjectMeta.ResourceVersion

			chartConfigsToUpdate = append(chartConfigsToUpdate, chartConfigToUpdate)

			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found chartconfig '%s' that has to be updated", desiredChartConfig.GetName()))
		}
	}

	c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d chartconfigs which have to be updated", len(chartConfigsToUpdate)))

	return chartConfigsToUpdate, nil
}
