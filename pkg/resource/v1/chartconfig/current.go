package chartconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	clusterGuestConfig, err := r.toClusterGuestConfigFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	clusterConfig, err := prepareClusterConfig(r.baseClusterConfig, clusterGuestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	g8sClient, err := r.guestClusterService.GetG8sClient(ctx, clusterConfig.ClusterID, clusterConfig.Domain.API)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	chartConfigList, err := g8sClient.CoreV1alpha1().ChartConfigs(v1.NamespaceSystem).List(v1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	chartConfigs := make([]*v1alpha1.ChartConfig, 0, len(chartConfigList.Items))

	for _, item := range chartConfigList.Items {
		chartConfigs = append(chartConfigs, &item)
	}

	return chartConfigs, nil
}
