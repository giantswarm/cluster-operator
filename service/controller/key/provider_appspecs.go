package key

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AWSAppSpecs returns apps installed only for AWS.
func AWSAppSpecs() []AppSpec {
	// Add any provider specific charts here.
	return []AppSpec{
		{
			App:             "cert-manager",
			Catalog:         "default",
			Chart:           "cert-manager-app",
			ClusterAPIOnly:  true,
			Namespace:       metav1.NamespaceSystem,
			UseUpgradeForce: true,
			Version:         "1.0.4",
		},
		{
			App:             "cluster-autoscaler",
			Catalog:         "default",
			Chart:           "cluster-autoscaler-app",
			Namespace:       metav1.NamespaceSystem,
			UseUpgradeForce: true,
			Version:         "1.1.2",
		},
		{
			App:             "external-dns",
			Catalog:         "default",
			Chart:           "external-dns-app",
			ClusterAPIOnly:  true,
			Namespace:       metav1.NamespaceSystem,
			UseUpgradeForce: true,
			Version:         "1.1.0",
		},
		{
			App:             "kiam",
			Catalog:         "default-test",
			Chart:           "kiam-app",
			ClusterAPIOnly:  true,
			Namespace:       metav1.NamespaceSystem,
			UseUpgradeForce: true,
			Version:         "1.0.3-045eedac92e1694ff197671f92f901f2ad48ed73",
		},
	}
}

// AzureAppSpecs returns apps installed only for Azure.
func AzureAppSpecs() []AppSpec {
	// Add any provider specific charts here.
	return []AppSpec{
		{
			App:             "external-dns",
			Catalog:         "default",
			Chart:           "external-dns-app",
			Namespace:       metav1.NamespaceSystem,
			UseUpgradeForce: true,
			Version:         "1.1.0",
		},
	}
}

// KVMAppSpecs returns apps installed only for KVM.
func KVMAppSpecs() []AppSpec {
	// Add any provider specific charts here.
	return []AppSpec{}
}
