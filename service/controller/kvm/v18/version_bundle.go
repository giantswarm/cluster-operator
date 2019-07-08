package v18

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "coredns",
				Description: "Update to 1.5.1. https://github.com/giantswarm/kubernetes-coredns/blob/master/CHANGELOG.md",
				Kind:        versionbundle.KindRemoved,
			},
			{
				Component:   "nginx-ingress-controller",
				Description: "Disabled migration logic now migration to helm chart is complete.",
				Kind:        versionbundle.KindRemoved,
			},
			{
				Component:   "node-exporter",
				Description: "Disable ipvs collector.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "node-exporter",
				Description: "Fix monitored file system mount points.",
				Kind:        versionbundle.KindFixed,
			},
			{
				Component:   "node-exporter",
				Description: "Fix systemd collector D-Bus connection. https://github.com/giantswarm/kubernetes-node-exporter/pull/44",
				Kind:        versionbundle.KindFixed,
			},
			{
				Component:   "metrics-server",
				Description: "Add network policy.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "net-exporter",
				Description: "Add network policy.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "coredns",
				Description: "Add network policy.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "kube-state-metrics",
				Description: "Add network policy.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "nginx-ingress-controller",
				Description: "Add network policy.",
				Kind:        versionbundle.KindAdded,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "coredns",
				Version: "1.5.1",
			},
			{
				Name:    "kube-state-metrics",
				Version: "1.5.0",
			},
			{
				Name:    "nginx-ingress-controller",
				Version: "0.24.1",
			},
			{
				Name:    "node-exporter",
				Version: "0.18.0",
			},
			{
				Name:    "metrics-server",
				Version: "0.3.1",
			},
		},
		Name:     "cluster-operator",
		Provider: "kvm",
		Version:  "0.17.0",
	}
}
