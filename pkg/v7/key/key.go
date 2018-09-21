package key

import (
	"fmt"
	"net"
	"strings"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/versionbundle"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// defaultDNSLastOctet is the last octect for the DNS service IP, the first
	// 3 octets come from the cluster IP range.
	defaultDNSLastOctet = 10
)

// APIAltNames returns the alt names for API certs.
func APIAltNames(clusterID string, kubeAltNames []string) []string {
	return append(kubeAltNames, fmt.Sprintf("master.%s", clusterID))
}

// APIDomain returns the API server domain for the guest cluster.
func APIDomain(clusterGuestConfig v1alpha1.ClusterGuestConfig) (string, error) {
	return serverDomain(clusterGuestConfig, certs.APICert)
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

// CIDRBlock returns a CIDR block for the given address and prefix.
func CIDRBlock(address, prefix string) string {
	if address == "" && prefix == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s", address, prefix)
}

// ClusterID returns cluster ID for given guest cluster config.
func ClusterID(clusterGuestConfig v1alpha1.ClusterGuestConfig) string {
	return clusterGuestConfig.ID
}

// ClusterOrganization returns the org for given guest cluster config.
func ClusterOrganization(clusterGuestConfig v1alpha1.ClusterGuestConfig) string {
	return clusterGuestConfig.Owner
}

// CommonChartSpecs returns charts installed for all providers.
// Note: When adding chart specs you also need to add the chart name to the
// desired state test in the chartconfig service.
func CommonChartSpecs() []ChartSpec {
	return []ChartSpec{
		{
			AppName:       "cert-exporter",
			ChannelName:   "stable",
			ChartName:     "cert-exporter-chart",
			ConfigMapName: "cert-exporter-values",
			Namespace:     metav1.NamespaceSystem,
			ReleaseName:   "cert-exporter",
		},
		{
			AppName:       "kube-state-metrics",
			ChannelName:   "0-1-stable",
			ChartName:     "kubernetes-kube-state-metrics-chart",
			ConfigMapName: "kube-state-metrics-values",
			Namespace:     metav1.NamespaceSystem,
			ReleaseName:   "kube-state-metrics",
		},
		{
			AppName:       "net-exporter",
			ChannelName:   "stable",
			ChartName:     "net-exporter-chart",
			ConfigMapName: "net-exporter-values",
			Namespace:     metav1.NamespaceSystem,
			ReleaseName:   "net-exporter",
		},
		{
			AppName:       "nginx-ingress-controller",
			ChannelName:   "0-2-stable",
			ChartName:     "kubernetes-nginx-ingress-controller-chart",
			ConfigMapName: "nginx-ingress-controller-values",
			Namespace:     metav1.NamespaceSystem,
			ReleaseName:   "nginx-ingress-controller",
		},
		{
			AppName:       "node-exporter",
			ChannelName:   "0-1-stable",
			ChartName:     "kubernetes-node-exporter-chart",
			ConfigMapName: "node-exporter-values",
			Namespace:     metav1.NamespaceSystem,
			ReleaseName:   "node-exporter",
		},
	}
}

// DNSIP returns the IP of the DNS service given a cluster IP range.
func DNSIP(clusterIPRange string) (string, error) {
	ip, _, err := net.ParseCIDR(clusterIPRange)
	if err != nil {
		return "", microerror.Maskf(invalidConfigError, err.Error())
	}

	// Only IPV4 CIDRs are supported.
	ip = ip.To4()
	if ip == nil {
		return "", microerror.Mask(invalidConfigError)
	}

	// IP must be a network address.
	if ip[3] != 0 {
		return "", microerror.Mask(invalidConfigError)
	}

	ip[3] = defaultDNSLastOctet

	return ip.String(), nil
}

// DNSZone returns common domain for guest cluster.
func DNSZone(clusterGuestConfig v1alpha1.ClusterGuestConfig) string {
	return clusterGuestConfig.DNSZone
}

// EncryptionKeySecretName generates name for a Kubernetes secret based on
// information in given v1alpha1.ClusterGuestConfig.
func EncryptionKeySecretName(clusterGuestConfig v1alpha1.ClusterGuestConfig) string {
	return fmt.Sprintf("%s-%s", ClusterID(clusterGuestConfig), "encryption")
}

// IsDeleted returns true if the Kubernetes resource has been marked for
// deletion.
func IsDeleted(objectMeta metav1.ObjectMeta) bool {
	return objectMeta.DeletionTimestamp != nil
}

// MasterServiceDomain returns the domain of the master service for the given
// guest cluster.
func MasterServiceDomain(clusterGuestConfig v1alpha1.ClusterGuestConfig) string {
	return fmt.Sprintf("master.%s", ClusterID(clusterGuestConfig))
}

// serverDomain returns the guest cluster domain for the provided cluster
// component.
func serverDomain(clusterGuestConfig v1alpha1.ClusterGuestConfig, cert certs.Cert) (string, error) {
	commonDomain := DNSZone(clusterGuestConfig)

	if !strings.Contains(commonDomain, ".") {
		return "", microerror.Maskf(invalidConfigError, "commonDomain must be a valid domain")
	}

	return string(cert) + "." + strings.TrimLeft(commonDomain, "\t ."), nil
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
