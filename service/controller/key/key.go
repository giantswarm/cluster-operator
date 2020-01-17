package key

import (
	"fmt"
	"net"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// defaultDNSLastOctet is the last octect for the DNS service IP, the first
	// 3 octets come from the cluster IP range.
	defaultDNSLastOctet = 10
)

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

// CertConfigCertOperatorVersion returns version bundle version for given
// CertConfig.
func CertConfigCertOperatorVersion(cr v1alpha1.CertConfig) string {
	if cr.Labels == nil {
		return ""
	}

	return cr.Labels[label.CertOperatorVersion]
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
			Version:         "1.4.3",
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
