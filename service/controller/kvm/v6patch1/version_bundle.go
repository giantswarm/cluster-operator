package v6patch1

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "kube-state-metrics",
				Description: "Updated to 1.5.0. More info here: https://github.com/kubernetes/kube-state-metrics/blob/v1.5.0/CHANGELOG.md",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "kube-state-metrics",
				Description: "Added addon resizer. More info https://github.com/kubernetes/autoscaler/blob/master/addon-resizer/README.md",
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
				Version: "0.15.0",
			},
			{
				Name:    "node-exporter",
				Version: "0.15.1",
			},
		},
		Name:     "cluster-operator",
		Provider: "kvm",
		Version:  "0.6.1",
	}
}
