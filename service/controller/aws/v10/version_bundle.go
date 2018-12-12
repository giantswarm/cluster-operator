package v10

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "<replace me>",
				Description: "<replace me>",
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
				Version: "0.21.0",
			},
			{
				Name:    "node-exporter",
				Version: "0.15.1",
			},
			{
				Name:    "coredns",
				Version: "1.1.1",
			},
		},
		Name:     "cluster-operator",
		Provider: "aws",
		Version:  "0.10.0",
	}
}
