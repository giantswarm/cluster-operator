package chartconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v20/controllercontext"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}
	chartConfigs, err := toChartConfigs(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(chartConfigs) > 0 {
		for _, chartConfig := range chartConfigs {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating chartconfig %#q in namespace %#q", chartConfig.Name, chartConfig.Namespace))

			_, err := cc.Client.TenantCluster.G8s.CoreV1alpha1().ChartConfigs(chartConfig.Namespace).Update(chartConfig)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated chartconfig %#q in namespace %#q", chartConfig.Name, chartConfig.Namespace))
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not update chartconfigs")
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	delete, err := r.newDeleteChangeForUpdatePatch(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetCreateChange(create)
	patch.SetDeleteChange(delete)
	patch.SetUpdateChange(update)

	return patch, nil
}

// newDeleteChangeForUpdatePatch is specific to the update behaviour because we
// might want to remove certain chart configs when a tenant cluster is
// reconciled. So the delete change computed here is gathered for the update
// patch above.
func (r *Resource) newDeleteChangeForUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) ([]*v1alpha1.ChartConfig, error) {
	currentChartConfigs, err := toChartConfigs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredChartConfigs, err := toChartConfigs(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var chartConfigsToDelete []*v1alpha1.ChartConfig

	for _, currentChartConfig := range currentChartConfigs {
		_, err := getChartConfigByName(desiredChartConfigs, currentChartConfig.Name)
		// Existing ChartConfig is not desired anymore so it should be deleted.
		if IsNotFound(err) {
			chartConfigsToDelete = append(chartConfigsToDelete, currentChartConfig)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return chartConfigsToDelete, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) ([]*v1alpha1.ChartConfig, error) {
	currentChartConfigs, err := toChartConfigs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredChartConfigs, err := toChartConfigs(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

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
		}
	}

	return chartConfigsToUpdate, nil
}
