package v13

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "nginx-ingress-controller",
				Description: "Updated to 0.23.0. More info here https://github.com/kubernetes/ingress-nginx/blob/master/Changelog.md. Note: This has a breaking change to rewrite rules in Ingress objects https://kubernetes.github.io/ingress-nginx/examples/rewrite/#rewrite-target.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "nginx-ingress-controller",
				Description: "Removed CPU and memory limits to Ingress Controller pods. As per upstream recommendation.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "nginx-ingress-controller",
				Description: "Added hard anti-affinity rule to run a maximum of one Ingress Controller pod per node. As per upstream recommendation.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "nginx-ingress-controller",
				Description: "Enabled dynamic certificates to avoid reloading due to certificates changes.",
				Kind:        versionbundle.KindAdded,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "nginx-ingress-controller",
				Version: "0.23.0",
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
		Version:  "0.13.0",
	}
}
