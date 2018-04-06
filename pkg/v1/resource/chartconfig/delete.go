package chartconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	clusterGuestConfig, err := r.toClusterGuestConfigFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	chartConfigsToCreate, err := toChartConfigs(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(chartConfigsToCreate) > 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting chartconfigs")

		clusterConfig, err := prepareClusterConfig(r.baseClusterConfig, clusterGuestConfig)
		if err != nil {
			return microerror.Mask(err)
		}

		g8sClient, err := r.guest.NewG8sClient(ctx, clusterConfig.ClusterID, clusterConfig.Domain.API)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, chartConfig := range chartConfigsToCreate {
			err := g8sClient.CoreV1alpha1().ChartConfigs(metav1.NamespaceSystem).Delete(chartConfig.Name, &metav1.DeleteOptions{})
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

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	delete, err := r.newDeleteChangeForDeletePatch(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
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
