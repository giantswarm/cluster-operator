package key

import (
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeytpr"
)

// ClusterID extracts clusterID from v1alpha1.KVMClusterConfig.
func ClusterID(customObject v1alpha1.KVMClusterConfig) string {
	return customObject.Spec.Guest.ID
}

// EncryptionKeySecretName generates name for a Kubernetes secret based on
// information in given v1alpha1.KVMClusterConfig.
func EncryptionKeySecretName(customObject v1alpha1.KVMClusterConfig) string {
	return fmt.Sprintf("%s-%s", ClusterID(customObject), randomkeytpr.EncryptionKey.String())
}

// ToCustomObject converts value to v1alpha1.KVMClusterConfig and returns it or
// error if type does not match.
func ToCustomObject(v interface{}) (v1alpha1.KVMClusterConfig, error) {
	customObjectPointer, ok := v.(*v1alpha1.KVMClusterConfig)
	if !ok {
		return v1alpha1.KVMClusterConfig{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.KVMClusterConfig{}, v)
	}
	customObject := *customObjectPointer

	return customObject, nil
}
