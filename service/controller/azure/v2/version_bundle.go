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
				Description: "Install chart-operator in kube-system namespace.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cluster-operator",
				Description: "Misleading component reference to azure-operator removed.",
				Kind:        versionbundle.KindFixed,
			},
		},
		Components:   []versionbundle.Component{},
		Dependencies: []versionbundle.Dependency{},
		Deprecated:   false,
		Name:         "cluster-operator",
		Provider:     "azure",
		Time:         time.Date(2018, time.April, 16, 12, 00, 0, 0, time.UTC),
		Version:      "0.2.0",
		WIP:          true,
	}
}
