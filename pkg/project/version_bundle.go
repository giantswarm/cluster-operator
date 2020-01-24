package project

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle(p string) versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "cluster-operator",
				Description: "Add resource implementation to update infrastructure reference labels.",
				Kind:        versionbundle.KindFixed,
				URLs: []string{
					"https://github.com/giantswarm/cluster-operator/pull/888",
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
			{
				Component:   "cluster-autoscaler",
				Description: "Adjust RBAC permissions.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/cluster-autoscaler-app/releases/tag/v1.1.3",
				},
			},
			{
				Component:   "cluster-autoscaler",
				Description: "Update to upstream version 1.16.2.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/cluster-autoscaler-app/releases/tag/v1.1.2",
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
				Component:   "external-dns",
				Description: "Add support AWS SDK configuration with explicit credentials.",
				Kind:        versionbundle.KindAdded,
				URLs: []string{
					"https://github.com/giantswarm/external-dns-app/releases/tag/v1.1.0",
				},
			},
			{
				Component:   "kube-state-metrics",
				Description: "Adjust RBAC permissions.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/kube-state-metrics-app/releases/tag/v1.0.2",
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
				Description: "Update dependencies to support Kubernetes 1.16.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/node-exporter-app/releases/tag/v1.1.1",
				},
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "kube-state-metrics",
				Version: "1.9.2",
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
		Provider: p,
		Version:  BundleVersion(),
	}
}
