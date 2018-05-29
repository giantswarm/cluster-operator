package v4

import (
	"time"

	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "cluster-operator",
				Description: "Add changes here.",
				Kind:        versionbundle.KindAdded,
			},
		},
		Components:   []versionbundle.Component{},
		Dependencies: []versionbundle.Dependency{},
		Deprecated:   false,
		Name:         "cluster-operator",
		Provider:     "aws",
		Time:         time.Date(2018, time.May, 28, 8, 21, 0, 0, time.UTC),
		Version:      "0.4.0",
		WIP:          true,
	}
}
