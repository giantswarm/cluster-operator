package chartconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	chartConfigsToCreate, err := toChartConfigs(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(chartConfigsToCreate) > 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating chartconfigs")

		guestG8sClient, err := r.getGuestG8sClient(ctx, obj)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, chartConfigToCreate := range chartConfigsToCreate {
			_, err := guestG8sClient.CoreV1alpha1().ChartConfigs(resourceNamespace).Create(chartConfigToCreate)
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "created chartconfigs")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no need to create chartconfigs")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) ([]*v1alpha1.ChartConfig, error) {
	currentChartConfigs, err := toChartConfigs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredChartConfigs, err := toChartConfigs(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out which chartconfigs have to be created")

	chartConfigsToCreate := make([]*v1alpha1.ChartConfig, 0)

	for _, desiredChartConfig := range desiredChartConfigs {
		if !containsChartConfig(currentChartConfigs, desiredChartConfig) {
			chartConfigsToCreate = append(chartConfigsToCreate, desiredChartConfig)
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d chartconfigs that have to be created", len(chartConfigsToCreate)))

	return chartConfigsToCreate, nil
}
