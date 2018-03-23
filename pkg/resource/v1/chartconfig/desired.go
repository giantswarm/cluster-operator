package chartconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/pkg/label"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	chartAPIVersion              = "core.giantswarm.io"
	chartKind                    = "ChartConfig"
	chartOperatorChart           = "quay.io/giantswarm/chart-operator-chart"
	chartOperatorChartConfigName = "chart-operator"
	chartOperatorChannel         = "stable"
	chartOperatorRelease         = "chart-operator"
)

// GetDesiredState returns all desired ChartConfigs for managed guest resources.
func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	desiredChartConfigs := make([]*v1alpha1.ChartConfig, 0)
	{
		chartOperatorChartConfig := newChartOperatorChartConfig(r.baseClusterConfig, r.projectName)
		desiredChartConfigs = append(desiredChartConfigs, chartOperatorChartConfig)
	}

	return desiredChartConfigs, nil
}

func newChartOperatorChartConfig(clusterConfig *cluster.Config, projectName string) *v1alpha1.ChartConfig {
	return &v1alpha1.ChartConfig{
		TypeMeta: apimetav1.TypeMeta{
			Kind:       chartKind,
			APIVersion: chartAPIVersion,
		},
		ObjectMeta: apimetav1.ObjectMeta{
			Name: chartOperatorChartConfigName,
			Labels: map[string]string{
				label.Cluster:      clusterConfig.ClusterID,
				label.ManagedBy:    projectName,
				label.Organization: clusterConfig.Organization,
			},
		},
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:    chartOperatorChart,
				Channel: chartOperatorChannel,
				Release: chartOperatorRelease,
			},
			VersionBundle: v1alpha1.ChartConfigSpecVersionBundle{
				Version: clusterConfig.VersionBundleVersion,
			},
		},
	}
}
