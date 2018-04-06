package chartconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	clusterGuestConfig, err := r.toClusterGuestConfigFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	chartConfigsToUpdate, err := toChartConfigs(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(chartConfigsToUpdate) > 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating chartconfigs")

		clusterConfig, err := prepareClusterConfig(r.baseClusterConfig, clusterGuestConfig)
		if err != nil {
			return microerror.Mask(err)
		}

		g8sClient, err := r.guest.NewG8sClient(ctx, clusterConfig.ClusterID, clusterConfig.Domain.API)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, chartConfigToUpdate := range chartConfigsToUpdate {
			_, err := g8sClient.CoreV1alpha1().ChartConfigs(metav1.NamespaceSystem).Update(chartConfigToUpdate)
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated chartconfigs")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no need to update chartconfigs")
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)

	return patch, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) ([]*v1alpha1.ChartConfig, error) {
	return []*v1alpha1.ChartConfig{}, nil
}
