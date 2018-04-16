package chartconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
)

const (
	chartConfigAPIVersion           = "core.giantswarm.io"
	chartConfigKind                 = "ChartConfig"
	chartConfigVersionBundleVersion = "0.1.0"
)

// GetDesiredState returns all desired ChartConfigs for managed guest resources.
func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	// TODO: Add ChartConfigs as we migrate components from k8scloudconfig.
	/*
		clusterGuestConfig, err := r.toClusterGuestConfigFunc(obj)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		clusterConfig, err := prepareClusterConfig(r.baseClusterConfig, clusterGuestConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	*/

	desiredChartConfigs := make([]*v1alpha1.ChartConfig, 0)

	return desiredChartConfigs, nil
}
