package key

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
)

// ToCustomObject converts value to v1alpha1.ClusterNetworkConfig and returns it or
// error if type does not match.
func ToCustomObject(v interface{}) (v1alpha1.ClusterNetworkConfig, error) {
	customObjectPointer, ok := v.(*v1alpha1.ClusterNetworkConfig)
	if !ok {
		return v1alpha1.ClusterNetworkConfig{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.ClusterNetworkConfig{}, v)
	}

	if customObjectPointer == nil {
		return v1alpha1.ClusterNetworkConfig{}, microerror.Maskf(emptyValueError,
			"empty value cannot be converted to CustomObject")
	}

	return *customObjectPointer, nil
}

// VersionBundleVersion extracts version bundle version from ClusterNetworkConfig.
func VersionBundleVersion(clusterNetworkConfig v1alpha1.ClusterNetworkConfig) string {
	return clusterNetworkConfig.Spec.VersionBundle.Version
}
