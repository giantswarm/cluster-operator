package key

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
)

// AWSConfigName generates name to AWSConfig CR from AWSClusterConfig.
func AWSConfigName(awsClusterConfig v1alpha1.AWSClusterConfig) string {
	return awsClusterConfig.Spec.Guest.ID
}

// ClusterGuestConfig extracts ClusterGuestConfig from AWSClusterConfig.
func ClusterGuestConfig(awsClusterConfig v1alpha1.AWSClusterConfig) v1alpha1.ClusterGuestConfig {
	return awsClusterConfig.Spec.Guest.ClusterGuestConfig
}

// ToCustomObject converts value to v1alpha1.AWSClusterConfig and returns it or
// error if type does not match.
func ToCustomObject(v interface{}) (v1alpha1.AWSClusterConfig, error) {
	customObjectPointer, ok := v.(*v1alpha1.AWSClusterConfig)
	if !ok {
		return v1alpha1.AWSClusterConfig{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.AWSClusterConfig{}, v)
	}

	if customObjectPointer == nil {
		return v1alpha1.AWSClusterConfig{}, microerror.Maskf(emptyValueError, "empty value cannot be converted to CustomObject")
	}

	return *customObjectPointer, nil
}

// VersionBundleVersion extracts version bundle version from AWSClusterConfig.
func VersionBundleVersion(awsClusterConfig v1alpha1.AWSClusterConfig) string {
	return awsClusterConfig.Spec.VersionBundle.Version
}

func WorkerCount(awsClusterConfig v1alpha1.AWSClusterConfig) int {
	return len(awsClusterConfig.Spec.Guest.Workers)
}
