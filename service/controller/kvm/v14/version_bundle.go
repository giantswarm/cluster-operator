package v14

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "coredns",
				Description: "Updated to 1.5.0. More info here: https://coredns.io/2019/04/06/coredns-1.5.0-release/",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "coredns",
				Version: "1.5.0",
			},
			{
				Name:    "kube-state-metrics",
				Version: "1.5.0",
			},
			{
				Name:    "nginx-ingress-controller",
				Version: "0.23.0",
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
		Version:  "0.14.0",
	}
}
