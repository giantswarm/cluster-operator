package project

import (
	"github.com/giantswarm/versionbundle"
)

var versionBundleAWS = versionbundle.Bundle{
	Changelogs: []versionbundle.Changelog{
		{
			Component:   "chart-operator",
			Description: "Adjust ClusterRole permissions.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/chart-operator/releases/tag/v0.11.3",
			},
		},
		{
			Component:   "cert-manager",
			Description: "Improve helm chart for clusters with restrictive network policies.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/cert-manager-app/releases/tag/v1.0.4",
			},
		},
		{
			Component:   "cert-manager",
			Description: "Update manifests for Kubernetes 1.16 compatibility.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/cert-manager-app/releases/tag/v1.0.3",
			},
		},
		{
			Component:   "kiam",
			Description: "Improve helm chart for clusters with restrictive network policies.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/kiam-app/releases/tag/v1.0.3",
			},
		},
		{
			Component:   "kiam",
			Description: "Update manifests for Kubernetes 1.16 compatibility.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/kiam-app/releases/tag/v1.0.2",
			},
		},
		{
			Component:   "metrics-server",
			Description: "Update manifests for Kubernetes 1.16 compatibility.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/metrics-server-app/releases/tag/v1.0.0",
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
			Name:    "cluster-autoscaler",
			Version: "1.16.2",
		},
		{
			Name:    "metrics-server",
			Version: "0.3.3",
		},
	},
	Name:     "cluster-operator",
	Provider: "aws",
	Version:  BundleVersion(),
}
