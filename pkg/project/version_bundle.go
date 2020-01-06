package project

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle(p string) versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "cluster-operator",
				Description: "Fix cluster status conditions to be reconciled upon cluster creation.",
				Kind:        versionbundle.KindFixed,
				URLs: []string{
					"TODO",
				},
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "kube-state-metrics",
				Version: "1.9.0",
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
		Version:  BundleVersion(),
	}
}
