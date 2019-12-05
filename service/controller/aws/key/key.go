package key

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/service/controller/key"
)

// AppSpecs returns apps installed only for AWS.
func AppSpecs() []key.AppSpec {
	// Add any provider specific charts here.
	return []key.AppSpec{
		{
			App:             "external-dns",
			Catalog:         "default",
			Chart:           "external-dns-app",
			ClusterAPIOnly:  true,
			Namespace:       metav1.NamespaceSystem,
			UseUpgradeForce: true,
			Version:         "1.0.0",
		},
		{
			App:             "kiam",
			Catalog:         "default",
			Chart:           "kiam-app",
			ClusterAPIOnly:  true,
			Namespace:       metav1.NamespaceSystem,
			UseUpgradeForce: true,
			Version:         "1.0.0",
		},
		{
			App:             "cert-manager",
			Catalog:         "default",
			Chart:           "cert-manager-app",
			ClusterAPIOnly:  true,
			Namespace:       metav1.NamespaceSystem,
			UseUpgradeForce: true,
			Version:         "1.0.0",
		},
		{
			App:             "cluster-autoscaler",
			Catalog:         "default",
			Chart:           "cluster-autoscaler-app",
			Namespace:       metav1.NamespaceSystem,
			UseUpgradeForce: true,
			Version:         "1.1.0",
		},
	}
}
