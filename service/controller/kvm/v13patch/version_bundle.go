package v13patch

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "nginx-ingress-controller",
				Description: "Updated to 0.24.1. More info here https://github.com/kubernetes/ingress-nginx/blob/master/Changelog.md",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "coredns",
				Version: "1.3.1",
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
				Version: "0.15.1",
			},
			{
				Name:    "metrics-server",
				Version: "0.3.1",
			},
		},
		Name:     "cluster-operator",
		Provider: "kvm",
		Version:  "0.13.1",
	}
}
