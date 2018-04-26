package chartconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	clusterGuestConfig, err := r.toClusterGuestConfigFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	chartConfigsToCreate, err := toChartConfigs(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(chartConfigsToCreate) > 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating chartconfigs")

		clusterConfig, err := prepareClusterConfig(r.baseClusterConfig, clusterGuestConfig)
		if err != nil {
			return microerror.Mask(err)
		}

		g8sClient, err := r.guest.NewG8sClient(ctx, clusterConfig.ClusterID, clusterConfig.Domain.API)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, chartConfigToCreate := range chartConfigsToCreate {
			_, err := g8sClient.CoreV1alpha1().ChartConfigs(metav1.NamespaceSystem).Create(chartConfigToCreate)
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
