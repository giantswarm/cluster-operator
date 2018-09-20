package chartconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/errors/guest"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/v7/key"
)

const (
	chartConfigAPIVersion           = "core.giantswarm.io"
	chartConfigKind                 = "ChartConfig"
	chartConfigVersionBundleVersion = "0.3.0"
)

func (c *ChartConfig) GetDesiredState(ctx context.Context, clusterConfig ClusterConfig, providerChartSpecs []key.ChartSpec) ([]*v1alpha1.ChartConfig, error) {
	desiredChartConfigs := make([]*v1alpha1.ChartConfig, 0)

	// Add any provider specific chart specs.
	chartSpecs := append(key.CommonChartSpecs(), providerChartSpecs...)

	for _, chartSpec := range chartSpecs {
		chartConfigCR, err := c.newChartConfig(ctx, clusterConfig, chartSpec)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		desiredChartConfigs = append(desiredChartConfigs, chartConfigCR)
	}

	return desiredChartConfigs, nil
}

func (c *ChartConfig) newChartConfig(ctx context.Context, clusterConfig ClusterConfig, chartSpec key.ChartSpec) (*v1alpha1.ChartConfig, error) {
	configMapSpec, err := c.newConfigMapSpec(ctx, clusterConfig, chartSpec)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	labels := newChartConfigLabels(clusterConfig, chartSpec.AppName, c.projectName)
	chartConfigCR := &v1alpha1.ChartConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       chartConfigKind,
			APIVersion: chartConfigAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   chartSpec.ChartName,
			Labels: labels,
		},
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:      chartSpec.ChartName,
				Namespace: chartSpec.Namespace,
				Channel:   chartSpec.ChannelName,
				ConfigMap: *configMapSpec,
				Release:   chartSpec.ReleaseName,
			},
			VersionBundle: v1alpha1.ChartConfigSpecVersionBundle{
				Version: chartConfigVersionBundleVersion,
			},
		},
	}

	return chartConfigCR, nil
}

func (c *ChartConfig) newConfigMapSpec(ctx context.Context, clusterConfig ClusterConfig, chartSpec key.ChartSpec) (*v1alpha1.ChartConfigSpecConfigMap, error) {
	if chartSpec.ConfigMapName == "" {
		// Return early. Nothing to do.
		return &v1alpha1.ChartConfigSpecConfigMap{}, nil
	}

	tenantK8sClient, err := c.newTenantK8sClient(ctx, clusterConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	configMap, err := tenantK8sClient.CoreV1().ConfigMaps(chartSpec.Namespace).Get(chartSpec.ConfigMapName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) || guest.IsAPINotAvailable(err) {
		// Cannot get configmap resource version so leave it unset. We will
		// check again after the next resync period.
		configMapSpec := &v1alpha1.ChartConfigSpecConfigMap{
			Name:      chartSpec.ConfigMapName,
			Namespace: chartSpec.Namespace,
		}

		return configMapSpec, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	// Set the configmap resource version. When this changes it will generate
	// an update event for chart-operator. chart-operator will recalculate the
	// desired state including any updated config map values.
	configMapSpec := &v1alpha1.ChartConfigSpecConfigMap{
		Name:            chartSpec.ConfigMapName,
		Namespace:       chartSpec.Namespace,
		ResourceVersion: configMap.ResourceVersion,
	}

	return configMapSpec, nil
}

func newChartConfigLabels(clusterConfig ClusterConfig, appName, projectName string) map[string]string {
	return map[string]string{
		label.App:          appName,
		label.Cluster:      clusterConfig.ClusterID,
		label.ManagedBy:    projectName,
		label.Organization: clusterConfig.Organization,
		label.ServiceType:  label.ServiceTypeManaged,
	}
}
