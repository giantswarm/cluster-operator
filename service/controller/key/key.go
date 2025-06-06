package key

import (
	"fmt"
	"net"

	"github.com/giantswarm/apiextensions/v6/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/cluster-operator/v5/pkg/label"
)

const (
	// defaultDNSLastOctet is the last octect for the DNS service IP, the first
	// 3 octets come from the cluster IP range.
	defaultDNSLastOctet = 10

	// UniqueOperatorVersion This is a special version used to indicate that the App CR
	// should be reconciled by the workload cluster app-operator.
	UniqueOperatorVersion = "0.0.0"
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

// DNSIP returns the IP of the DNS service given a cluster IP range.
func DNSIP(clusterIPRange string) (string, error) {
	ip, _, err := net.ParseCIDR(clusterIPRange)
	if err != nil {
		return "", microerror.Maskf(invalidConfigError, "%s", err.Error())
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
