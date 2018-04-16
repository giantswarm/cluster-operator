package v1

import (
	"time"

	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "Cluster Operator",
				Description: "Initial version for KVM",
				Kind:        "added",
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "kvm-operator",
				Version: "1.0.0",
			},
		},
		Dependencies: []versionbundle.Dependency{},
		Deprecated:   false,
		Name:         "cluster-operator",
		Provider:     "kvm",
		Time:         time.Date(2018, time.April, 16, 11, 00, 0, 0, time.UTC),
		Version:      "0.1.0",
		WIP:          false,
	}
}
