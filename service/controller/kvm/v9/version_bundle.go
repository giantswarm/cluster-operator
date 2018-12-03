package v9

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "nginx-ingress-controller",
				Description: "Update to version 0.21.0",
				Kind:        versionbundle.KindAdded,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "coredns",
				Version: "1.1.1",
			},
			{
				Name:    "kube-state-metrics",
				Version: "1.3.1",
			},
			{
				Name:    "nginx-ingress-controller",
				Version: "0.21.0",
			},
			{
				Name:    "node-exporter",
				Version: "0.15.1",
			},
		},
		Name:     "cluster-operator",
		Provider: "kvm",
		Version:  "0.9.0",
	}
}
