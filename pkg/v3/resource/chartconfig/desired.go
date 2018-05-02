package chartconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/pkg/label"
)

const (
	chartConfigAPIVersion           = "core.giantswarm.io"
	chartConfigKind                 = "ChartConfig"
	chartConfigVersionBundleVersion = "0.1.0"
)

// GetDesiredState returns all desired ChartConfigs for managed guest resources.
func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	clusterGuestConfig, err := r.toClusterGuestConfigFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterConfig, err := prepareClusterConfig(r.baseClusterConfig, clusterGuestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredChartConfigs := make([]*v1alpha1.ChartConfig, 0)
	{
		chartConfig := newKubeStateMetricsMigrationChartConfig(clusterConfig, r.projectName)
		desiredChartConfigs = append(desiredChartConfigs, chartConfig)
	}
	{
		chartConfig := newKubeStateMetricsChartConfig(clusterConfig, r.projectName)
		desiredChartConfigs = append(desiredChartConfigs, chartConfig)
	}

	return desiredChartConfigs, nil
}

func newKubeStateMetricsChartConfig(clusterConfig cluster.Config, projectName string) *v1alpha1.ChartConfig {
	return &v1alpha1.ChartConfig{
		TypeMeta: apimetav1.TypeMeta{
			Kind:       chartConfigKind,
			APIVersion: chartConfigAPIVersion,
		},
		ObjectMeta: apimetav1.ObjectMeta{
			Name: "kubernetes-kube-state-metrics-chart",
			Labels: map[string]string{
				label.Cluster:      clusterConfig.ClusterID,
				label.ManagedBy:    projectName,
				label.Organization: clusterConfig.Organization,
			},
		},
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:      "kubernetes-kube-state-metrics-chart",
				Channel:   "0-1-beta",
				Namespace: "kube-system",
				Release:   "kube-state-metrics",
			},
			VersionBundle: v1alpha1.ChartConfigSpecVersionBundle{
				Version: clusterConfig.VersionBundleVersion,
			},
		},
	}
}

// newKubeStateMetricsMigrationChartConfig adds the migration chart for
// kube-state-metrics. It will be removed in the next release once existing
// clusters have been upgraded.
func newKubeStateMetricsMigrationChartConfig(clusterConfig cluster.Config, projectName string) *v1alpha1.ChartConfig {
	return &v1alpha1.ChartConfig{
		TypeMeta: apimetav1.TypeMeta{
			Kind:       chartConfigKind,
			APIVersion: chartConfigAPIVersion,
		},
		ObjectMeta: apimetav1.ObjectMeta{
			Name: "kubernetes-kube-state-metrics-migration-chart",
			Labels: map[string]string{
				label.Cluster:      clusterConfig.ClusterID,
				label.ManagedBy:    projectName,
				label.Organization: clusterConfig.Organization,
			},
		},
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:      "kubernetes-kube-state-metrics-migration-chart",
				Channel:   "0-1-beta",
				Namespace: "kube-system",
				Release:   "kube-state-metrics-migration",
			},
			VersionBundle: v1alpha1.ChartConfigSpecVersionBundle{
				Version: clusterConfig.VersionBundleVersion,
			},
		},
	}
}
