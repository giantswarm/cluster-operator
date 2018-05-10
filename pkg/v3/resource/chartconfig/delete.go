package chartconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	chartConfigsToCreate, err := toChartConfigs(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(chartConfigsToCreate) > 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting chartconfigs")

		guestG8sClient, err := r.getGuestG8sClient(ctx, obj)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, chartConfig := range chartConfigsToCreate {
			err := guestG8sClient.CoreV1alpha1().ChartConfigs(metav1.NamespaceSystem).Delete(chartConfig.Name, &metav1.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deleted chartconfigs")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no need to delete chartconfigs")
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	delete, err := r.newDeleteChangeForDeletePatch(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) newDeleteChangeForDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) ([]*v1alpha1.ChartConfig, error) {
	currentChartConfigs, err := toChartConfigs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d chartconfigs that have to be deleted", len(currentChartConfigs)))

	return currentChartConfigs, nil
}

func (r *Resource) newDeleteChangeForUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) ([]*v1alpha1.ChartConfig, error) {
	currentChartConfigs, err := toChartConfigs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredChartConfigs, err := toChartConfigs(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out which chartconfigs have to be deleted")

	chartConfigsToDelete := make([]*v1alpha1.ChartConfig, 0)

	for _, currentChartConfig := range currentChartConfigs {
		_, err := getChartConfigByName(desiredChartConfigs, currentChartConfig.Name)
		// Existing ChartConfig is not desired anymore so it should be deleted.
		if IsNotFound(err) {
			chartConfigsToDelete = append(chartConfigsToDelete, currentChartConfig)
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d chartconfigs that have to be deleted", len(chartConfigsToDelete)))

	return chartConfigsToDelete, nil
}
