package key

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
)

func ClusterID(azureClusterConfig v1alpha1.AzureClusterConfig) string {
	return azureClusterConfig.Spec.Guest.ID
}

func ClusterOrganization(azureClusterConfig v1alpha1.AzureClusterConfig) string {
	return azureClusterConfig.Spec.Guest.Owner
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
