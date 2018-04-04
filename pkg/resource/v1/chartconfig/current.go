package chartconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetCurrentState returns the ChartConfig resources present in the guest
// cluster.
func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	clusterGuestConfig, err := r.toClusterGuestConfigFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "looking for chartconfigs in the guest cluster")

	clusterConfig, err := prepareClusterConfig(r.baseClusterConfig, clusterGuestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	g8sClient, err := r.guestClusterService.GetG8sClient(ctx, clusterConfig.ClusterID, clusterConfig.Domain.API)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	chartConfigList, err := g8sClient.CoreV1alpha1().ChartConfigs(metav1.NamespaceSystem).List(metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	chartConfigs := make([]*v1alpha1.ChartConfig, 0, len(chartConfigList.Items))

	for _, item := range chartConfigList.Items {
		chartConfigs = append(chartConfigs, &item)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d chartconfigs in the guest cluster", len(chartConfigs)))

	return chartConfigs, nil
}
