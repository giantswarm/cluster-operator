package app

import "github.com/giantswarm/cluster-operator/service/controller/key"

var (
	Default = key.AppSpec{
		Catalog:         "default",
		Namespace:       "kube-system",
		UseUpgradeForce: true,
	}

	ConfigExceptions = map[string]key.AppSpec{
		"cert-exporter": {
			Chart: "cert-exporter",
		},
		// chart-operator must be installed first so the chart CRD is
		// created in the tenant cluster.
		"chart-operator": {
			Chart:     "chart-operator",
			Namespace: "giantswarm",
		},
		// CoreDNS's Upgrade force is disabled to avoid affecting customer workloads.
		"coredns": {
			UseUpgradeForce: false,
		},
		"net-exporter": {
			Chart: "net-exporter",
		},
	}
)
