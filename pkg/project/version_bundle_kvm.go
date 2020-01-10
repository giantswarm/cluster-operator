package project

import (
	"github.com/giantswarm/versionbundle"
)

var versionBundleKVM = versionbundle.Bundle{
	Changelogs: []versionbundle.Changelog{
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
			Component:   "chart-operator",
			Description: "Removed CPU limits.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https: //github.com/giantswarm/chart-operator/pull/335",
			},
		},
		{
			Component:   "coredns",
			Description: "Updated to version 1.6.5.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/coredns-app/blob/master/CHANGELOG.md#v113-2020-01-08",
			},
		},
		{
			Component:   "kube-state-metrics",
			Description: "Updated to version 1.9.0.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/kube-state-metrics-app/blob/master/CHANGELOG.md#v100",
			},
		},
		{
			Component:   "net-exporter",
			Description: "Updated to version 1.4.3.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/net-exporter/blob/master/CHANGELOG.md#143-2019-12-27",
			},
		},
		{
			Component:   "nginx-ingress-controller",
			Description: "Migrated to use default app catalog.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/nginx-ingress-controller-app/blob/master/CHANGELOG.md#v111-2020-01-04",
			},
		},
		{
			Component:   "node-exporter",
			Description: "Updated to version 0.18.1.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/node-exporter-app/blob/master/CHANGELOG.md#120-2020-01-08",
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
			Name:    "metrics-server",
			Version: "0.4.1",
		},
	},
	Name:     "cluster-operator",
	Provider: "kvm",
	Version:  BundleVersion(),
}
