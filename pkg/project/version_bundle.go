package project

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle(p string) versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "cluster-operator",
				Description: "Stop setting IC replicas count",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/cluster-operator/pull/949",
				},
			},
		},
		Components: []versionbundle.Component{},
		Name:       "cluster-operator",
		Provider:   p,
		Version:    Version(),
	}
}
