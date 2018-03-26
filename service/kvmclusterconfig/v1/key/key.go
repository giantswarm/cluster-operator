package key

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
)

// ClusterID extracts clusterID from v1alpha1.KVMClusterConfig.
func ClusterID(customObject v1alpha1.KVMClusterConfig) string {
	return customObject.Spec.Guest.ID
}

// ToCustomObject converts value to v1alpha1.KVMClusterConfig and returns it or
// error if type does not match.
func ToCustomObject(v interface{}) (v1alpha1.KVMClusterConfig, error) {
	customObjectPointer, ok := v.(*v1alpha1.KVMClusterConfig)
	if !ok {
		return v1alpha1.KVMClusterConfig{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.KVMClusterConfig{}, v)
	}

	if customObjectPointer == nil {
		return v1alpha1.KVMClusterConfig{}, microerror.Maskf(emptyValueError,
			"empty value cannot be converted to CustomObject")
	}

	return *customObjectPointer, nil
}
