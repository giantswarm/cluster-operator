package key

import (
	"fmt"
)

// CertAltNames returns the alt names for the Kubernetes API certs.
func CertAltNames(altNames ...string) []string {
	list := []string{
		"kubernetes",
		"kubernetes.default",
		"kubernetes.default.svc",
		"kubernetes.default.svc.cluster.local",
	}

	return append(list, altNames...)
}

// CertConfigName constructs a name for CertConfig CRs using the clusterI D and
// the cert name.
func CertConfigName(getter LabelsGetter, name string) string {
	return fmt.Sprintf("%s-%s", ClusterID(getter), name)
}
