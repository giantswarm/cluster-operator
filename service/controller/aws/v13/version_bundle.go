package v13

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "nginx-ingress-controller",
				Description: "Updated to 0.23.0. More info here https://github.com/kubernetes/ingress-nginx/blob/master/Changelog.md.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "nginx-ingress-controller",
				Description: "Removed CPU and memory limits to Ingress Controller pods as per upstream recommendation.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "nginx-ingress-controller",
				Description: "Run single Ingress Controller pod per worker node.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "nginx-ingress-controller",
				Description: "Enabled dynamic certificates flag.",
				Kind:        versionbundle.KindAdded,
			},
		},
		Components: []versionbundle.Component{
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
				Name:    "coredns",
				Version: "1.3.1",
			},
			{
				Name:    "cluster-autoscaler",
				Version: "1.3.1",
			},
			{
				Name:    "metrics-server",
				Version: "0.3.1",
			},
		},
		Name:     "cluster-operator",
		Provider: "aws",
		Version:  "0.13.0",
	}
}
