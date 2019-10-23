package v21

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "coredns",
				Description: "Migrated to use default app catalog.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "kube-state-metrics",
				Description: "Migrated to use default app catalog.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "metrics-server",
				Description: "Migrated to use default app catalog.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "node-exporter",
				Description: "Migrated to use default app catalog.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "chart-operator",
				Description: "Extended to support multiple DNS servers when bootstrapping coredns.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "kube-state-metrics",
				Description: "Updated to v1.8.0. https://github.com/giantswarm/kube-state-metrics-app/blob/72d4d3804b9fd8c46e1f23013ed2d78efed5ecca/CHANGELOG.md#v060",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "nginx-ingress-controller",
				Description: "Updated to v0.26.1. https://github.com/giantswarm/kubernetes-nginx-ingress-controller/blob/d0616c69eb224d49cdfd9a9b63e7cf61d59335ae/CHANGELOG.md#100",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "coredns",
				Description: "Updated to v1.16.4. https://github.com/giantswarm/kubernetes-coredns/blob/eb9e6979e4bb35b03bd9b0c99e1c68b6a4f3b4c6/CHANGELOG.md#v080",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "metrics-server",
				Description: "Updated to v0.4.1. https://github.com/giantswarm/metrics-server-app/blob/db61c5c13c4ba55c357ab098253b3f3fe8f4e2cd/CHANGELOG.md#v041",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "kube-state-metrics",
				Version: "1.8.0",
			},
			{
				Name:    "nginx-ingress-controller",
				Version: "0.26.1",
			},
			{
				Name:    "node-exporter",
				Version: "0.18.0",
			},
			{
				Name:    "coredns",
				Version: "1.6.4",
			},
			{
				Name:    "metrics-server",
				Version: "0.4.1",
			},
		},
		Name:     "cluster-operator",
		Provider: "kvm",
		Version:  "0.21.0",
	}
}
