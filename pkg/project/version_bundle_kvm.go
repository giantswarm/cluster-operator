package project

import (
	"github.com/giantswarm/versionbundle"
)

var versionBundleKVM = versionbundle.Bundle{
	Changelogs: []versionbundle.Changelog{
		{
			Component:   "kube-state-metrics",
			Description: "Updated to upstream kube-state-metrics v1.9.0.",
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
