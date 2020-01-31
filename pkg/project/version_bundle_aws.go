package project

import (
	"github.com/giantswarm/versionbundle"
)

var versionBundleAWS = versionbundle.Bundle{
	Changelogs: []versionbundle.Changelog{
		{
			Component:   "nginx-ingress-controller",
			Description: "Supported enabling the proxy protocol with AWS ELB.",
			Kind:        versionbundle.KindAdded,
			URLs: []string{
				"https://github.com/giantswarm/nginx-ingress-controller-app/blob/master/CHANGELOG.md#v121-2020-01-29",
			},
		},
		{
			Component:   "nginx-ingress-controller",
			Description: "Updated to version 0.28.0.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/nginx-ingress-controller-app/blob/master/CHANGELOG.md#v130-2020-01-30",
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
			Version: "0.28.0",
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
