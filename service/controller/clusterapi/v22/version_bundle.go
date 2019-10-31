package v22

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle(p string) versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "nodepools",
				Description: "Add Node Pools functionality. See https://docs.giantswarm.io/basics/nodepools/ for details.",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "kube-state-metrics",
				Version: "1.8.0",
			},
			{
				Name:    "nginx-ingress-controller",
				Version: "0.26.1",
			},
			{
				Name:    "node-exporter",
				Version: "0.18.0",
			},
			{
				Name:    "coredns",
				Version: "1.6.4",
			},
			{
				Name:    "cluster-autoscaler",
				Version: "1.15.2",
			},
			{
				Name:    "metrics-server",
				Version: "0.4.1",
			},
		},
		Name:     "cluster-operator",
		Provider: p,
		Version:  "1.0.0",
	}
}
