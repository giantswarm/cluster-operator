package v12

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "cluster-operator",
				Description: "Reduce installation time of managed services in new tenant clusters.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "coredns",
				Description: "Updated to 1.3.1. More info here: https://coredns.io/2019/01/13/coredns-1.3.1-release/",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "coredns",
				Description: "Added a custom block parameter to allow advanced configuration",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "nginx-ingress-controller",
				Version: "0.21.0",
			},
			{
				Name:    "external-dns",
				Version: "0.5.2",
			},
			{
				Name:    "kube-state-metrics",
				Version: "1.5.0",
			},
			{
				Name:    "node-exporter",
				Version: "0.15.1",
			},
			{
				Name:    "coredns",
				Version: "1.3.1",
			},
			{
				Name:    "metrics-server",
				Version: "0.3.1",
			},
		},
		Name:     "cluster-operator",
		Provider: "azure",
		Version:  "0.11.0",
	}
}
