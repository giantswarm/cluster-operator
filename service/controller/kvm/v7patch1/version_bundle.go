package v7patch1

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "chart-operator",
				Description: "Added support for user configmaps.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "metrics-server",
				Description: "Added metrics-server managed service.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "nginx-ingress-controller",
				Description: "Made ingress-nginx configmap configurable.",
				Kind:        versionbundle.KindAdded,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "kube-state-metrics",
				Version: "1.3.1",
			},
			{
				Name:    "nginx-ingress-controller",
				Version: "0.15.0",
			},
			{
				Name:    "node-exporter",
				Version: "0.15.1",
			},
		},
		Name:     "cluster-operator",
		Provider: "kvm",
		Version:  "0.7.0",
	}
}
