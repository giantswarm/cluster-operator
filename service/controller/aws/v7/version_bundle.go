package v7

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "chart-operator",
				Description: "Added monitoring for managed charts.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "cluster-operator",
				Description: "Added support for deploying cert-exporter.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "cluster-operator",
				Description: "Added support for deploying net-exporter.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "cluster-operator",
				Description: "Added support for deploying nginx-ingress-controller.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "nginx-ingress-controller",
				Description: "Updated to 0.15.0.",
				Kind:        versionbundle.KindChanged,
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
		Provider: "aws",
		Version:  "0.6.0",
	}
}
