package key

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/service/controller/key"
)

// AppSpecs returns apps installed only for Azure.
func AppSpecs() []key.AppSpec {
	// Add any provider specific charts here.
	return []key.AppSpec{
		{
			App:             "external-dns",
			Catalog:         "default",
			Chart:           "external-dns-app",
			Namespace:       metav1.NamespaceSystem,
			UseUpgradeForce: true,
			Version:         "1.0.0",
		},
	}
}
