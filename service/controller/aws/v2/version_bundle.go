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
				Description: "Removed misleading component reference to aws-operator.",
				Kind:        versionbundle.KindFixed,
			},
			{
				Component:   "cluster-operator",
				Description: "Removed chart resource so Tiller is not installed in the kube-system namespace.",
				Kind:        versionbundle.KindRemoved,
			},
		},
		Components: []versionbundle.Component{},
		Name:       "cluster-operator",
		Provider:   "aws",
		Version:    "0.2.0",
	}
}
