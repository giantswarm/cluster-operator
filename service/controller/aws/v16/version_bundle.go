package v16

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "cluster-operator",
				Description: "Added support for creating a cluster configmap for use by managed apps.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "nginx-ingress-controller",
				Description: "Added separate podsecuritypolicy.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "nginx-ingress-controller",
				Description: "Switched security context to non-root user.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "coredns",
				Description: "Added separate podsecuritypolicy.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "coredns",
				Description: "Switched security context to non-root user.",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{
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
				Name:    "coredns",
				Version: "1.5.0",
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
		Version:  "0.16.0",
	}
}
