package key

import (
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/versionbundle"
)

// APIEndpoint returns guest cluster Kubernetes API endpoint.
func APIEndpoint(clusterGuestConfig v1alpha1.ClusterGuestConfig) string {
	return clusterGuestConfig.API.Endpoint
}

// CertConfigName constructs a name for CertConfig CR using ClusterID and Cert.
func CertConfigName(clusterID string, cert certs.Cert) string {
	return fmt.Sprintf("%s-%s", clusterID, cert)
}

// CertConfigVersionBundleVersion returns version bundle version for given
// CertConfig.
func CertConfigVersionBundleVersion(customObject v1alpha1.CertConfig) string {
	return customObject.Spec.VersionBundle.Version
}

// ClusterID returns cluster ID for given guest cluster config.
func ClusterID(clusterGuestConfig v1alpha1.ClusterGuestConfig) string {
	return clusterGuestConfig.ID
}

// VersionBundles returns slice of versionbundle.Bundles for given guest
// cluster config.
func VersionBundles(clusterGuestConfig v1alpha1.ClusterGuestConfig) []versionbundle.Bundle {
	versionBundles := make([]versionbundle.Bundle, len(clusterGuestConfig.VersionBundles))
	for i, vb := range clusterGuestConfig.VersionBundles {
		versionBundles[i].Name = vb.Name
		versionBundles[i].Version = vb.Version
	}

	return versionBundles
}
