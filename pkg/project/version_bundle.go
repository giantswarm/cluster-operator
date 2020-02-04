package project

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle(p string) versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "net-exporter",
				Description: "Update to 1.6.0.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/net-exporter/releases/tag/v1.6.0",
				},
			},
			{
				Component:   "kiam",
				Description: "Update to 3.5.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/kiam-app/releases/tag/v1.1.0",
				},
			},
			{
				Component:   "cert-manager",
				Description: "Update to 0.13.0.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/cert-manager-app/releases/tag/v1.1.0",
				},
			},
			{
				Component:   "external-dns",
				Description: "Update to 0.5.18.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/external-dns-app/releases/tag/v1.2.0",
				},
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "kube-state-metrics",
				Version: "1.9.2",
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
				Version: "0.3.3",
			},
			{
				Name:    "kiam",
				Version: "3.5.0",
			},
			{
				Name:    "external-dns",
				Version: "0.5.18",
			},
			{
				Name:    "cert-manager",
				Version: "0.13.0",
			},
			{
				Name:    "net-exporter",
				Version: "1.6.0",
			},
		},
		Name:     "cluster-operator",
		Provider: p,
		Version:  BundleVersion(),
	}
}
