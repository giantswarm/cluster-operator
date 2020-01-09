package project

import (
	"github.com/giantswarm/versionbundle"
)

var versionBundleKVM = versionbundle.Bundle{
	Changelogs: []versionbundle.Changelog{
		{
			Component:   "coredns",
			Description: "Updated to version 1.6.5.",
			Kind:        versionbundle.KindChanged,
		},
		{
			Component:   "kube-state-metrics",
			Description: "Updated to version 1.9.0.",
			Kind:        versionbundle.KindChanged,
		},
		{
			Component:   "net-exporter",
			Description: "Updated to version 1.4.3.",
			Kind:        versionbundle.KindChanged,
		},
		{
			Component:   "node-exporter",
			Description: "Updated to version 0.18.1.",
			Kind:        versionbundle.KindChanged,
		},
		{
			Component:   "nginx-ingress-controller",
			Description: "Migrated to use default app catalog.",
			Kind:        versionbundle.KindChanged,
		},
		{
			Component:   "cluster-operator",
			Description: "Added additional settings for coredns to cluster configmap.",
			Kind:        versionbundle.KindChanged,
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
