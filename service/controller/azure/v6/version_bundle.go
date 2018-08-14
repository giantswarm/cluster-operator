package v6

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
		},
		Components: []versionbundle.Component{
			{
				Name:    "nginx-ingress-controller",
				Version: "0.12.0",
			},
			{
				Name:    "external-dns",
				Version: "0.5.2",
			},
			{
				Name:    "kube-state-metrics",
				Version: "1.3.1",
			},
			{
				Name:    "node-exporter",
				Version: "0.15.1",
			},
		},
		Name:     "cluster-operator",
		Provider: "azure",
		Version:  "0.6.0",
	}
}
