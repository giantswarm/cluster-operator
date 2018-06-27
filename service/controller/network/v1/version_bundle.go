package v1

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "cluster-operator-network",
				Description: "Initial version for cluster-operator network controller",
				Kind:        versionbundle.KindAdded,
			},
		},
		Components: []versionbundle.Component{},
		Name:       "cluster-operator-network",
		Version:    "0.1.0",
	}
}
