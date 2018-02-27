package key

import (
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/certs"
)

func CertConfigName(clusterGuestConfig v1alpha1.ClusterGuestConfig, cert certs.Cert) string {
	return fmt.Sprintf("%s-%s", ClusterID(clusterGuestConfig), cert)
}

func ClusterID(clusterGuestConfig v1alpha1.ClusterGuestConfig) string {
	return clusterGuestConfig.ID
}
