package v22

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "nodepools",
				Description: "Add Node Pools functionality. See https://docs.giantswarm.io/basics/nodepools/ for details.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "kiam",
				Description: "Add managed kiam app into default app catalog(aws only).",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "external-dns",
				Description: "Add managed external-dns app into default app catalog.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "cert-manager",
				Description: "Add managed cert-manager app into default app catalog.",
				Kind:        versionbundle.KindAdded,
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
				Name:    "cluster-autoscaler",
				Version: "1.15.2",
			},
			{
				Name:    "metrics-server",
				Version: "0.4.1",
			},
			{
				Name:    "kiam",
				Version: "3.4.0",
			},
			{
				Name:    "external-dns",
				Version: "0.5.11",
			},
			{
				Name:    "cert-manager",
				Version: "0.9.0",
			},
		},
		Name:    "cluster-operator",
		Version: "0.22.0",
	}
}
