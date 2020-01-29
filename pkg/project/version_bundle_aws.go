package project

import (
	"github.com/giantswarm/versionbundle"
)

var versionBundleAWS = versionbundle.Bundle{
	Changelogs: []versionbundle.Changelog{
		{
			Component:   "cluster-operator",
			Description: "TODO.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"",
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
			Name:    "cluster-autoscaler",
			Version: "1.16.2",
		},
		{
			Name:    "metrics-server",
			Version: "0.3.3",
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
		{
			Name:    "net-exporter",
			Version: "1.6.0",
		},
	},
	Name:     "cluster-operator",
	Provider: "aws",
	Version:  BundleVersion(),
}
