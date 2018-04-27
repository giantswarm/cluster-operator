package v1

import (
	"time"

	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "cluster-operator",
				Description: "Initial version for Azure",
				Kind:        versionbundle.KindAdded,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "azure-operator",
				Version: "1.0.0",
			},
		},
		Dependencies: []versionbundle.Dependency{},
		Deprecated:   true,
		Name:         "cluster-operator",
		Provider:     "azure",
		Time:         time.Date(2018, time.March, 28, 7, 30, 0, 0, time.UTC),
		Version:      "0.1.0",
		WIP:          false,
	}
}
