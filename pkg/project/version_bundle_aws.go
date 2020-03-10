package project

import (
	"github.com/giantswarm/versionbundle"
)

var versionBundleAWS = versionbundle.Bundle{
	Changelogs: []versionbundle.Changelog{
		{
			Component:   "cluster-operator",
			Description: "Support extra small cluster profile detection",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/cluster-operator/pull/956",
			},
		},
	},
	Components: []versionbundle.Component{},
	Name:       "cluster-operator",
	Provider:   "aws",
	Version:    BundleVersion(),
}
