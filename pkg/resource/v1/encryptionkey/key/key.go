package key

import (
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
)

// ClusterID extracts clusterID from v1alpha1.ClusterGuestConfig.
func ClusterID(clusterGuestConfig v1alpha1.ClusterGuestConfig) string {
	return clusterGuestConfig.ID
}

// EncryptionKeySecretName generates name for a Kubernetes secret based on
// information in given v1alpha1.ClusterGuestConfig.
func EncryptionKeySecretName(clusterGuestConfig v1alpha1.ClusterGuestConfig) string {
	return fmt.Sprintf("%s-%s", ClusterID(clusterGuestConfig), "encryption")
}
