package v18

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "cert-exporter",
				Description: "Add toleration for all taints",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "net-exporter",
				Description: "Add toleration for all taints",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "node-exporter",
				Description: "Add toleration for all taints",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "nginx-ingress-controller",
				Description: "Update to 0.25.0. https://github.com/giantswarm/kubernetes-nginx-ingress-controller/blob/master/CHANGELOG.md",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "coredns",
				Version: "1.5.1",
			},
			{
				Name:    "kube-state-metrics",
				Version: "1.5.0",
			},
			{
				Name:    "nginx-ingress-controller",
				Version: "0.25.0",
			},
			{
				Name:    "node-exporter",
				Version: "0.18.0",
			},
			{
				Name:    "metrics-server",
				Version: "0.3.1",
			},
		},
		Name:     "cluster-operator",
		Provider: "kvm",
		Version:  "0.18.0",
	}
}
