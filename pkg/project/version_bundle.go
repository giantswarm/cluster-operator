package project

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle(p string) versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "cluster-operator",
				Description: "Moved default app list from cluster-operator code to release repository.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/cluster-operator/pull/889",
				},
			},
		},
		Components: []versionbundle.Component{},
		Name:       "cluster-operator",
		Provider:   p,
		Version:    BundleVersion(),
	}
}
