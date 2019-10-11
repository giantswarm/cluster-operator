package v21

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "kube-state-metrics",
				Description: "Migrated to use default app catalog.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "metrics-server",
				Description: "Migrated to use default app catalog.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "node-exporter",
				Description: "Migrated to use default app catalog.",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "kube-state-metrics",
				Version: "1.7.2",
			},
			{
				Name:    "nginx-ingress-controller",
				Version: "0.25.1",
			},
			{
				Name:    "node-exporter",
				Version: "0.18.0",
			},
			{
				Name:    "coredns",
				Version: "1.6.2",
			},
			{
				Name:    "cluster-autoscaler",
				Version: "1.14.0",
			},
			{
				Name:    "metrics-server",
				Version: "0.3.1",
			},
		},
		Name:     "cluster-operator",
		Provider: "aws",
		Version:  "0.21.0",
	}
}
