package v2

import (
	"time"

	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "Cluster Operator",
				Description: "TODO",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "azure-operator",
				Version: "1.0.0",
			},
		},
		Dependencies: []versionbundle.Dependency{},
		Deprecated:   false,
		Name:         "cluster-operator",
		Provider:     "azure",
		Time:         time.Date(2018, time.April, 16, 12, 00, 0, 0, time.UTC),
		Version:      "0.2.0",
		WIP:          true,
	}
}
