package key

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/versionbundle"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	IngressControllerConfigMapName = "ingress-controller-values"

	// defaultDNSLastOctet is the last octect for the DNS service IP, the first
	// 3 octets come from the cluster IP range.
	defaultDNSLastOctet = 10
)

// APIAltNames returns the alt names for API certs.
func APIAltNames(clusterID, internalAPI string, kubeAltNames []string) []string {
	additionalAltNames := []string{
		fmt.Sprintf("master.%s", clusterID),
		internalAPI,
	}

	return append(kubeAltNames, additionalAltNames...)
}

// APIDomain returns the API server domain for the guest cluster.
func APIDomain(clusterGuestConfig v1alpha1.ClusterGuestConfig) (string, error) {
	return serverDomain(clusterGuestConfig, certs.APICert)
}

// AppUserConfigMapName returns the name of the user values configmap for the
// given app spec.
func AppUserConfigMapName(appSpec AppSpec) string {
	return fmt.Sprintf("%s-user-values", appSpec.App)
}

// AppUserSecretName returns the name of the user values secret for the
// given app spec.
func AppUserSecretName(appSpec AppSpec) string {
	return fmt.Sprintf("%s-user-secrets", appSpec.App)
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

// ClusterConfigMapName returns the cluster name used in the configMap generated for this tenant cluster.
func ClusterConfigMapName(clusterGuestConfig v1alpha1.ClusterGuestConfig) string {
	return fmt.Sprintf("%s-cluster-values", clusterGuestConfig.ID)
}

// ClusterID returns cluster ID for given guest cluster config.
func ClusterID(clusterGuestConfig v1alpha1.ClusterGuestConfig) string {
	return clusterGuestConfig.ID
}

// ClusterOrganization returns the org for given guest cluster config.
func ClusterOrganization(clusterGuestConfig v1alpha1.ClusterGuestConfig) string {
	return clusterGuestConfig.Owner
}

// CommonAppSpecs returns apps installed for all providers.
func CommonAppSpecs() []AppSpec {
	return []AppSpec{
		{
			// chart-operator must be installed first so the chart CRD is
			// created in the tenant cluster.
			App:             "chart-operator",
			Catalog:         "default",
			Chart:           "chart-operator",
			Namespace:       "giantswarm",
			UseUpgradeForce: true,
			Version:         "0.11.3",
		},
		{
			// coredns must be installed second as its a requirement for other
			// apps.
			App:       "coredns",
			Catalog:   "default",
			Chart:     "coredns-app",
			Namespace: metav1.NamespaceSystem,
			// Upgrade force is disabled to avoid affecting customer workloads.
			UseUpgradeForce: false,
			Version:         "1.1.3",
		},
		{
			App:             "cert-exporter",
			Catalog:         "default",
			Chart:           "cert-exporter",
			Namespace:       metav1.NamespaceSystem,
			UseUpgradeForce: true,
			Version:         "1.2.1",
		},
		{
			App:             "kube-state-metrics",
			Catalog:         "default",
			Chart:           "kube-state-metrics-app",
			Namespace:       metav1.NamespaceSystem,
			UseUpgradeForce: true,
			Version:         "1.0.0",
		},
		{
			App:             "metrics-server",
			Catalog:         "default",
			Chart:           "metrics-server-app",
			Namespace:       metav1.NamespaceSystem,
			UseUpgradeForce: true,
			Version:         "1.0.0",
		},
		{
			App:             "net-exporter",
			Catalog:         "default",
			Chart:           "net-exporter",
			Namespace:       metav1.NamespaceSystem,
			UseUpgradeForce: true,
			Version:         "1.5.1",
		},
		{
			App:           "nginx-ingress-controller",
			Catalog:       "default",
			Chart:         "nginx-ingress-controller-app",
			ConfigMapName: IngressControllerConfigMapName,
			Namespace:     metav1.NamespaceSystem,
			// Upgrade force is disabled to avoid dropping customer traffic
			// that is using the Ingress Controller.
			UseUpgradeForce: false,
			Version:         "1.1.1",
		},
		{
			App:             "node-exporter",
			Catalog:         "default",
			Chart:           "node-exporter-app",
			Namespace:       metav1.NamespaceSystem,
			UseUpgradeForce: true,
			Version:         "1.2.0",
		},
	}
}

// CommonChartSpecs returns charts installed for all providers.
// Note: When adding chart specs you also need to add the chart name to the
// desired state tests in the chartconfig and configmap services.
func CommonChartSpecs() []ChartSpec {
	return []ChartSpec{
		{
			AppName:           "coredns",
			ChartName:         "kubernetes-coredns-chart",
			ConfigMapName:     "coredns-values",
			UserConfigMapName: "coredns-user-values",
		},
		{
			AppName:       "cert-exporter",
			ChartName:     "cert-exporter-chart",
			ConfigMapName: "cert-exporter-values",
		},
		{
			AppName:       "kube-state-metrics",
			ChartName:     "kubernetes-kube-state-metrics-chart",
			ConfigMapName: "kube-state-metrics-values",
		},
		{
			AppName:       "metrics-server",
			ChartName:     "kubernetes-metrics-server-chart",
			ConfigMapName: "metrics-server-values",
		},
		{
			AppName:       "net-exporter",
			ChartName:     "net-exporter-chart",
			ConfigMapName: "net-exporter-values",
		},
		{
			AppName:           "nginx-ingress-controller",
			ChartName:         "kubernetes-nginx-ingress-controller-chart",
			ConfigMapName:     "nginx-ingress-controller-values",
			UserConfigMapName: "nginx-ingress-controller-user-values",
		},
		{
			AppName:       "node-exporter",
			ChartName:     "kubernetes-node-exporter-chart",
			ConfigMapName: "node-exporter-values",
		},
	}
}

// CordonUntilDate sets the date that chartconfig CRs should be cordoned until
// when they are migrated to app CRs.
func CordonUntilDate() string {
	return time.Now().Add(1 * time.Hour).Format("2006-01-02T15:04:05")
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

// KubeConfigClusterName returns the cluster name used in the kubeconfig generated for this tenant cluster.
func KubeConfigClusterName(clusterGuestConfig v1alpha1.ClusterGuestConfig) string {
	return fmt.Sprintf("giantswarm-%s", clusterGuestConfig.ID)
}

// KubeConfigSecretName returns the name of secret resource for a tenant cluster
func KubeConfigSecretName(clusterGuestConfig v1alpha1.ClusterGuestConfig) string {
	return fmt.Sprintf("%s-kubeconfig", clusterGuestConfig.ID)
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
