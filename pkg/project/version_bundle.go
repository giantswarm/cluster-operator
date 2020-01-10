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
					"https://github.com/giantswarm/cluster-operator/pull/866",
				},
			},
			{
				Component:   "cluster-operator",
				Description: "Added additional settings for coredns to cluster configmap.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/cluster-operator/pull/871",
				},
			},
			{
				Component:   "cert-exporter",
				Description: "Removed CPU limits.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/cert-exporter/blob/master/CHANGELOG.md#121-2019-12-24",
				},
			},
			{
				Component:   "cert-manager",
				Description: "Removed CPU limits.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/cert-manager-app/blob/master/CHANGELOG.md#v103-2020-01-03",
				},
			},
			{
				Component:   "chart-operator",
				Description: "Removed CPU limits.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https: //github.com/giantswarm/chart-operator/pull/335",
				},
			},
			{
				Component:   "cluster-autoscaler",
				Description: "Removed CPU limits.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/cluster-autoscaler-app/blob/master/CHANGELOG.md#v112-2020-01-03",
				},
			},
			{
				Component:   "external-dns",
				Description: "Added support AWS SDK configuration with explicit credentials.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/external-dns-app/blob/master/CHANGELOG.md#v110",
				},
			},
			{
				Component:   "external-dns",
				Description: "Removed CPU limits.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/external-dns-app/blob/master/CHANGELOG.md#v110",
				},
			},
			{
				Component:   "kiam",
				Description: "Removed CPU limits.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/kiam-app/blob/master/CHANGELOG.md#v102-2020-01-04",
				},
			},
			{
				Component:   "node-exporter",
				Description: "Changed priority class to system-node-critical",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/node-exporter-app/blob/master/CHANGELOG.md#120-2020-01-08",
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
