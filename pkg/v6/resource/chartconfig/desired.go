package chartconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/errors/guest"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/pkg/label"
)

const (
	chartConfigAPIVersion           = "core.giantswarm.io"
	chartConfigKind                 = "ChartConfig"
	chartConfigVersionBundleVersion = "0.2.0"
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
		chartConfig, err := r.newKubeStateMetricsChartConfig(ctx, clusterConfig, r.projectName)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		desiredChartConfigs = append(desiredChartConfigs, chartConfig)
	}
	{
		chartConfig, err := r.newNodeExporterChartConfig(ctx, clusterConfig, r.projectName)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		desiredChartConfigs = append(desiredChartConfigs, chartConfig)
	}

	// Enable Ingress Controller for Azure and AWS.
	if r.provider == label.ProviderAzure || r.provider == label.ProviderAWS {
		chartConfig, err := r.newIngressControllerChartConfig(ctx, clusterConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		desiredChartConfigs = append(desiredChartConfigs, chartConfig)
	}

	// Only enable External DNS for Azure.
	if r.provider == label.ProviderAzure {
		chartConfig := newExternalDNSChartConfig(clusterConfig, r.projectName)
		desiredChartConfigs = append(desiredChartConfigs, chartConfig)
	}

	return desiredChartConfigs, nil
}

func (r *Resource) newIngressControllerChartConfig(ctx context.Context, clusterConfig cluster.Config) (*v1alpha1.ChartConfig, error) {
	chartName := "kubernetes-nginx-ingress-controller-chart"
	channelName := "0-2-stable"
	configMapName := "nginx-ingress-controller-values"
	releaseName := "nginx-ingress-controller"
	labels := newChartConfigLabels(clusterConfig, releaseName, r.projectName)

	configMapSpec, err := r.getConfigMapSpec(ctx, clusterConfig, configMapName, apismetav1.NamespaceSystem)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	chartConfigCR := &v1alpha1.ChartConfig{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       chartConfigKind,
			APIVersion: chartConfigAPIVersion,
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name:   chartName,
			Labels: labels,
		},
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:      chartName,
				Namespace: apismetav1.NamespaceSystem,
				Channel:   channelName,
				ConfigMap: *configMapSpec,
				Release:   releaseName,
			},
			VersionBundle: v1alpha1.ChartConfigSpecVersionBundle{
				Version: chartConfigVersionBundleVersion,
			},
		},
	}

	return chartConfigCR, nil
}

func newExternalDNSChartConfig(clusterConfig cluster.Config, projectName string) *v1alpha1.ChartConfig {
	chartName := "kubernetes-external-dns-chart"
	channelName := "0-1-stable"
	releaseName := "external-dns"
	labels := newChartConfigLabels(clusterConfig, releaseName, projectName)

	return &v1alpha1.ChartConfig{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       chartConfigKind,
			APIVersion: chartConfigAPIVersion,
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name:   chartName,
			Labels: labels,
		},
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:      chartName,
				Channel:   channelName,
				Namespace: apismetav1.NamespaceSystem,
				Release:   releaseName,
			},
			VersionBundle: v1alpha1.ChartConfigSpecVersionBundle{
				Version: chartConfigVersionBundleVersion,
			},
		},
	}
}

func (r *Resource) newKubeStateMetricsChartConfig(ctx context.Context, clusterConfig cluster.Config, projectName string) (*v1alpha1.ChartConfig, error) {
	chartName := "kubernetes-kube-state-metrics-chart"
	channelName := "0-1-stable"
	configMapName := "kube-state-metrics-values"
	releaseName := "kube-state-metrics"
	labels := newChartConfigLabels(clusterConfig, releaseName, projectName)

	configMapSpec, err := r.getConfigMapSpec(ctx, clusterConfig, configMapName, apismetav1.NamespaceSystem)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	chartConfigCR := &v1alpha1.ChartConfig{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       chartConfigKind,
			APIVersion: chartConfigAPIVersion,
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name:   chartName,
			Labels: labels,
		},
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:      chartName,
				Channel:   channelName,
				ConfigMap: *configMapSpec,
				Namespace: apismetav1.NamespaceSystem,
				Release:   releaseName,
			},
			VersionBundle: v1alpha1.ChartConfigSpecVersionBundle{
				Version: chartConfigVersionBundleVersion,
			},
		},
	}
	return chartConfigCR, nil
}

func (r *Resource) newNodeExporterChartConfig(ctx context.Context, clusterConfig cluster.Config, projectName string) (*v1alpha1.ChartConfig, error) {
	chartName := "kubernetes-node-exporter-chart"
	channelName := "0-1-stable"
	configMapName := "node-exporter-values"
	releaseName := "node-exporter"
	labels := newChartConfigLabels(clusterConfig, releaseName, projectName)

	configMapSpec, err := r.getConfigMapSpec(ctx, clusterConfig, configMapName, apismetav1.NamespaceSystem)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	chartConfigCR := &v1alpha1.ChartConfig{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       chartConfigKind,
			APIVersion: chartConfigAPIVersion,
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name:   chartName,
			Labels: labels,
		},
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:      chartName,
				Channel:   channelName,
				ConfigMap: *configMapSpec,
				Namespace: apismetav1.NamespaceSystem,
				Release:   releaseName,
			},
			VersionBundle: v1alpha1.ChartConfigSpecVersionBundle{
				Version: chartConfigVersionBundleVersion,
			},
		},
	}
	return chartConfigCR, nil
}

func (r *Resource) getConfigMapSpec(ctx context.Context, guestConfig cluster.Config, configMapName, namespace string) (*v1alpha1.ChartConfigSpecConfigMap, error) {
	guestK8sClient, err := r.getGuestK8sClient(ctx, guestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	configMap, err := guestK8sClient.CoreV1().ConfigMaps(namespace).Get(configMapName, apismetav1.GetOptions{})
	if apierrors.IsNotFound(err) || guest.IsAPINotAvailable(err) {
		// Cannot get configmap resource version so leave it unset. We will
		// check again after the next resync period.
		configMapSpec := &v1alpha1.ChartConfigSpecConfigMap{
			Name:      configMapName,
			Namespace: namespace,
		}

		return configMapSpec, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	// Set the configmap resource version. When this changes it will generate
	// an update event for chart-operator. chart-operator will recalculate the
	// desired state including any updated config map values.
	configMapSpec := &v1alpha1.ChartConfigSpecConfigMap{
		Name:            configMapName,
		Namespace:       namespace,
		ResourceVersion: configMap.ResourceVersion,
	}

	return configMapSpec, nil
}

func newChartConfigLabels(clusterConfig cluster.Config, appName, projectName string) map[string]string {
	return map[string]string{
		label.App:          appName,
		label.Cluster:      clusterConfig.ClusterID,
		label.ManagedBy:    projectName,
		label.Organization: clusterConfig.Organization,
		label.ServiceType:  label.ServiceTypeManaged,
	}
}
