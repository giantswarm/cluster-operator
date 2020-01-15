package project

import (
	"github.com/giantswarm/versionbundle"
)

var versionBundleKVM = versionbundle.Bundle{
	Changelogs: []versionbundle.Changelog{
		{
			Component:   "nginx-ingress-controller",
			Description: "Updated to version 0.27.1.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/nginx-ingress-controller-app/blob/master/CHANGELOG.md#v120-2020-01-13",
			},
		},
	},
	Components: []versionbundle.Component{
		{
			Name:    "kube-state-metrics",
			Version: "1.9.2",
		},
		{
			Name:    "nginx-ingress-controller",
			Version: "0.27.1",
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
			Version: "0.3.3",
		},
	},
	Name:     "cluster-operator",
	Provider: "kvm",
	Version:  BundleVersion(),
}
