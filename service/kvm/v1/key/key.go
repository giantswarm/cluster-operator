package key

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
)

func ToCustomObject(v interface{}) (v1alpha1.KVMClusterConfig, error) {
	customObjectPointer, ok := v.(*v1alpha1.KVMClusterConfig)
	if !ok {
		return v1alpha1.KVMClusterConfig{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.KVMClusterConfig{}, v)
	}
	customObject := *customObjectPointer

	return customObject, nil
}

func VersionBundleVersion(customObject v1alpha1.KVMClusterConfig) string {
	// return customObject.Spec.Guest.VersionBundles.Version
	return "unknown"
}
