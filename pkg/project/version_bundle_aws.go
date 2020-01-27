package project

import (
	"github.com/giantswarm/versionbundle"
)

var versionBundleAWS = versionbundle.Bundle{
	Changelogs: []versionbundle.Changelog{
		{
			Component:   "nginx-ingress-controller",
			Description: "Updated to version 0.27.1.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/nginx-ingress-controller-app/blob/master/CHANGELOG.md#v120-2020-01-13",
			},
		},
		{
			Component:   "nginx-ingress-controller",
			Description: "Added metrics Service for prometheus-operator support.",
			Kind:        versionbundle.KindAdded,
			URLs: []string{
				"https://github.com/giantswarm/nginx-ingress-controller-app/blob/master/CHANGELOG.md#v120-2020-01-13",
			},
		},
		{
			Component:   "nginx-ingress-controller",
			Description: "Supported configuring which SSL/TLS protocols should be enabled.",
			Kind:        versionbundle.KindAdded,
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
	},
	Name:     "cluster-operator",
	Provider: "aws",
	Version:  BundleVersion(),
}
