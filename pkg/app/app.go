package app

import "github.com/giantswarm/cluster-operator/service/controller/key"

var (
	Default = key.AppSpec{
		Catalog:         "default",
		Namespace:       "kube-system",
		UseUpgradeForce: true,
	}

	Exceptions = map[string]key.AppSpec{
		"cert-exporter": {
			Chart: "cert-exporter",
		},
		"chart-operator": {
			// chart-operator must be installed first so the chart CRD is
			// created in the tenant cluster.
			Chart:     "chart-operator",
			Namespace: "giantswarm",
		},
		"coredns": {
			// Upgrade force is disabled to avoid affecting customer workloads.
			UseUpgradeForce: false,
		},
		"net-exporter": {
			Chart: "net-exporter",
		},
	}
)
