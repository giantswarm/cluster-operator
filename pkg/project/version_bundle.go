package project

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle(p string) versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "cluster-operator",
				Description: "Add resource implementation to update infrastructure reference labels.",
				Kind:        versionbundle.KindFixed,
				URLs: []string{
					"https://github.com/giantswarm/cluster-operator/pull/888",
				},
			},
		},
		Components: []versionbundle.Component{},
		Name:       "cluster-operator",
		Provider:   p,
		Version:    BundleVersion(),
	}
}
