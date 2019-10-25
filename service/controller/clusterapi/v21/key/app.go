package key

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pkgkey "github.com/giantswarm/cluster-operator/pkg/v21/key"
)

// AppSpecs returns all app CRs for AWS for testing Node Pools.
//
// *****************************************************************************
// *** TODO: Remove and revert to provider specfic key packages once testing ***
// *** is complete.															 ***
// *****************************************************************************
//
func AppSpecs() []pkgkey.AppSpec {
	return []pkgkey.AppSpec{
		{
			App:             "cert-exporter",
			Catalog:         "default",
			Chart:           "cert-exporter",
			Namespace:       metav1.NamespaceSystem,
			UseUpgradeForce: true,
			Version:         "1.2.0",
		},
		{
			App:             "cluster-autoscaler",
			Catalog:         "default",
			Chart:           "cluster-autoscaler-app",
			Namespace:       metav1.NamespaceSystem,
			UseUpgradeForce: true,
			Version:         "0.10.0",
		},
		{
			App:       "coredns",
			Catalog:   "default",
			Chart:     "coredns-app",
			Namespace: metav1.NamespaceSystem,
			// Upgrade force is disabled to avoid affecting customer workloads.
			UseUpgradeForce: false,
			Version:         "0.9.0",
		},
		{
			App:             "chart-operator",
			Catalog:         "default",
			Chart:           "chart-operator",
			Namespace:       "giantswarm",
			UseUpgradeForce: true,
			Version:         "0.10.8",
		},
		{
			App:             "kube-state-metrics",
			Catalog:         "default",
			Chart:           "kube-state-metrics-app",
			Namespace:       metav1.NamespaceSystem,
			UseUpgradeForce: true,
			Version:         "0.6.0",
		},
		{
			App:             "metrics-server",
			Catalog:         "default",
			Chart:           "metrics-server-app",
			Namespace:       metav1.NamespaceSystem,
			UseUpgradeForce: true,
			Version:         "0.4.1",
		},
		{
			App:             "net-exporter",
			Catalog:         "default",
			Chart:           "net-exporter",
			Namespace:       metav1.NamespaceSystem,
			UseUpgradeForce: true,
			Version:         "1.3.0",
		},
		{
			App:             "node-exporter",
			Catalog:         "default",
			Chart:           "node-exporter-app",
			Namespace:       metav1.NamespaceSystem,
			UseUpgradeForce: true,
			Version:         "0.6.0",
		},
	}
}
