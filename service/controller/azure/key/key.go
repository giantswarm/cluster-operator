package key

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/service/controller/key"
)

// AppSpecs returns apps installed only for Azure.
func AppSpecs() []key.AppSpec {
	// Add any provider specific charts here.
	return []key.AppSpec{
		{
			App:             "external-dns",
			Catalog:         "default",
			Chart:           "external-dns-app",
			Namespace:       metav1.NamespaceSystem,
			UseUpgradeForce: true,
			Version:         "1.1.0",
		},
	}
}

// ChartSpecs returns charts installed only for Azure.
func ChartSpecs() []key.ChartSpec {
	return []key.ChartSpec{
		{
			AppName:   "external-dns",
			ChartName: "kubernetes-external-dns-chart",
		},
	}
}

// ClusterGuestConfig extracts ClusterGuestConfig from AzureClusterConfig.
func ClusterGuestConfig(azureClusterConfig v1alpha1.AzureClusterConfig) v1alpha1.ClusterGuestConfig {
	return azureClusterConfig.Spec.Guest.ClusterGuestConfig
}

// ToCustomObject converts value to v1alpha1.AzureClusterConfig and returns it or
// error if type does not match.
func ToCustomObject(v interface{}) (v1alpha1.AzureClusterConfig, error) {
	customObjectPointer, ok := v.(*v1alpha1.AzureClusterConfig)
	if !ok {
		return v1alpha1.AzureClusterConfig{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.AzureClusterConfig{}, v)
	}

	if customObjectPointer == nil {
		return v1alpha1.AzureClusterConfig{}, microerror.Maskf(emptyValueError, "empty value cannot be converted to CustomObject")
	}

	return *customObjectPointer, nil
}

func VersionBundleVersion(azureClusterConfig v1alpha1.AzureClusterConfig) string {
	return azureClusterConfig.Spec.VersionBundle.Version
}

func WorkerCount(azureClusterConfig v1alpha1.AzureClusterConfig) int {
	return len(azureClusterConfig.Spec.Guest.Workers)
}
