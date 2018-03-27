package key

import (
	"strings"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
)

// ClusterID returns cluster ID for given guest cluster config.
func ClusterID(clusterGuestConfig v1alpha1.ClusterGuestConfig) string {
	return clusterGuestConfig.ID
}

// DNSZone returns common domain for guest cluster.
func DNSZone(clusterGuestConfig v1alpha1.ClusterGuestConfig) string {
	return clusterGuestConfig.DNSZone
}

// ServerDomain returns the guest cluster domain for the provided cluster
// component.
func ServerDomain(clusterGuestConfig v1alpha1.ClusterGuestConfig, cert certs.Cert) (string, error) {
	commonDomain := DNSZone(clusterGuestConfig)

	if !strings.Contains(commonDomain, ".") {
		return "", microerror.Maskf(invalidConfigError, "commonDomain must be a valid domain")
	}

	return string(cert) + "." + strings.TrimLeft(commonDomain, "\t ."), nil
}
