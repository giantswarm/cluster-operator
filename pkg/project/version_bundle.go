package project

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle(p string) versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "cluster-operator",
				Description: "TODO",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"",
				},
			},
		},
		Components: []versionbundle.Component{},
		Name:       "cluster-operator",
		Provider:   p,
		Version:    Version(),
	}
}
