package project

import (
	"github.com/giantswarm/versionbundle"
)

var versionBundleAWS = versionbundle.Bundle{
	Changelogs: []versionbundle.Changelog{
		{
			Component:   "cluster-operator",
			Description: "Add additional settings for coredns to cluster configmap.",
			Kind:        versionbundle.KindAdded,
			URLs: []string{
				"https://github.com/giantswarm/cluster-operator/pull/873",
			},
		},
		{
			Component:   "chart-operator",
			Description: "Adjust ClusterRole permissions.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/chart-operator/releases/tag/v0.11.3",
			},
		},
		{
			Component:   "chart-operator",
			Description: "Remove CPU limits.",
			Kind:        versionbundle.KindRemoved,
			URLs: []string{
				"https://github.com/giantswarm/chart-operator/releases/tag/v0.11.2",
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
			Component:   "cert-manager",
			Description: "Remove CPU limits.",
			Kind:        versionbundle.KindRemoved,
			URLs: []string{
				"https://github.com/giantswarm/cert-manager-app/releases/tag/v1.0.2",
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
			Component:   "kiam",
			Description: "Remove CPU limits.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/kiam-app/releases/tag/v1.0.1",
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
		{
			Component:   "cert-exporter",
			Description: "Remove CPU limits.",
			Kind:        versionbundle.KindRemoved,
			URLs: []string{
				"https://github.com/giantswarm/cert-exporter/releases/tag/v1.2.1",
			},
		},
		{
			Component:   "cluster-autoscaler",
			Description: "Update to v1.16.2.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/cluster-autoscaler-app/releases/tag/v1.1.2",
			},
		},
		{
			Component:   "cluster-autoscaler",
			Description: "Remove CPU limits.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/cluster-autoscaler-app/releases/tag/v1.1.1",
			},
		},
		{
			Component:   "coredns",
			Description: "Update manifests for Kubernetes 1.16 compatibility.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/coredns-app/releases/tag/v1.1.3",
				"https://github.com/giantswarm/coredns-app/releases/tag/v1.1.2",
			},
		},
		{
			Component:   "coredns",
			Description: "Remove CPU limits.",
			Kind:        versionbundle.KindRemoved,
			URLs: []string{
				"https://github.com/giantswarm/coredns-app/releases/tag/v1.1.1",
			},
		},
		{
			Component:   "external-dns",
			Description: "Add support AWS SDK configuration with explicit credentials.",
			Kind:        versionbundle.KindAdded,
			URLs: []string{
				"https://github.com/giantswarm/external-dns-app/releases/tag/v1.1.0",
			},
		},
		{
			Component:   "external-dns",
			Description: "Remove CPU limits.",
			Kind:        versionbundle.KindRemoved,
			URLs: []string{
				"https://github.com/giantswarm/external-dns-app/releases/tag/v1.0.1",
			},
		},
		{
			Component:   "kube-state-metrics",
			Description: "Update to upstream version 1.9.2.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/kube-state-metrics-app/releases/tag/v1.0.1",
			},
		},
		{
			Component:   "net-exporter",
			Description: "Change priority class to system-node-critical.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/net-exporter/releases/tag/v1.5.1",
				"https://github.com/giantswarm/net-exporter/releases/tag/v1.5.0",
			},
		},
		{
			Component:   "node-exporter",
			Description: "Change priority class to system-node-critical.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/node-exporter-app/releases/tag/v1.2.0",
			},
		},
		{
			Component:   "node-exporter",
			Description: "Update dependencies to support Kubernetes 1.16.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/node-exporter-app/releases/tag/v1.1.1",
			},
		},
		{
			Component:   "node-exporter",
			Description: "Update to upstream version 0.18.1.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/node-exporter-app/releases/tag/v1.1.0",
			},
		},
		{
			Component:   "nginx-ingress-controller",
			Description: "Update manifests for Kubernetes 1.16.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/nginx-ingress-controller-app/releases/tag/v1.1.1",
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
