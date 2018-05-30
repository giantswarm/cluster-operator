package v2

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "cluster-operator",
				Description: "Installed chart-operator in kube-system namespace.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cluster-operator",
				Description: "Removed misleading component reference to kvm-operator.",
				Kind:        versionbundle.KindFixed,
			},
		},
		Components: []versionbundle.Component{},
		Name:       "cluster-operator",
		Provider:   "kvm",
		Version:    "0.2.0",
	}
}
