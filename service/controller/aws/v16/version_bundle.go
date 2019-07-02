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
			{
				Component:   "coredns",
				Description: "Default container port changed from 53 to 1053. Please read project's changelog carefully https://github.com/giantswarm/kubernetes-coredns/blob/master/CHANGELOG.md#v050",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "node-exporter",
				Description: "Added separate podsecuritypolicy.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "node-exporter",
				Description: "Switched security context to non-root user.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "node-exporter",
				Description: "Use force when doing helm upgrades to fix failed releases.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "metrics-server",
				Description: "Added separate podsecuritypolicy.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "metrics-server",
				Description: "Switched security context to non-root user.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "metrics-server",
				Description: "Use force when doing helm upgrades to fix failed releases.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "kube-state-metrics",
				Description: "Added separate podsecuritypolicy.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "kube-state-metrics",
				Description: "Switched security context to non-root user.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "kube-state-metrics",
				Description: "Use force when doing helm upgrades to fix failed releases.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "cluster-autoscaler",
				Description: "Added separate podsecuritypolicy.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "cluster-autoscaler",
				Description: "Switched security context to non-root user.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cluster-autoscaler",
				Description: "Use force when doing helm upgrades to fix failed releases.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "node-exporter",
				Description: "Updated to v0.18.0",
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
				Version: "0.18.0",
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
