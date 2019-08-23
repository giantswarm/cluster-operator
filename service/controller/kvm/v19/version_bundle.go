package v19

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "kube-state-metrics",
				Description: "Update to 1.7.2. https://github.com/giantswarm/kubernetes-kube-state-metrics/blob/master/CHANGELOG.md#v040",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "coredns",
				Description: "Update to 1.6.2. https://github.com/giantswarm/kubernetes-coredns/blob/master/CHANGELOG.md#v070",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "coredns",
				Version: "1.6.2",
			},
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
				Name:    "metrics-server",
				Version: "0.3.1",
			},
		},
		Name:     "cluster-operator",
		Provider: "kvm",
		Version:  "0.19.0",
	}
}
