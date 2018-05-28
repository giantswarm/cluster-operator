package v2

import (
	"time"

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
		},
		Components:   []versionbundle.Component{},
		Dependencies: []versionbundle.Dependency{},
		Deprecated:   true,
		Name:         "cluster-operator",
		Provider:     "aws",
		Time:         time.Date(2018, time.April, 16, 12, 00, 0, 0, time.UTC),
		Version:      "0.2.0",
		WIP:          false,
	}
}
