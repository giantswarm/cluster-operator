package v22

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle(p string) versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "cluster-autoscaler",
				Description: "Updated to version 1.16.2.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "coredns",
				Description: "Updated to version 1.6.5.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "net-exporter",
				Description: "Updated to version 1.4.0.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "node-exporter",
				Description: "Updated to version 0.18.1.",
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
				Version: "0.18.1",
			},
			{
				Name:    "coredns",
				Version: "1.6.5",
			},
			{
				Name:    "cluster-autoscaler",
				Version: "1.16.2",
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
		Name:     "cluster-operator",
		Provider: p,
		Version:  "1.0.0",
	}
}
