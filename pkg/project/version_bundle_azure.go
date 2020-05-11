package project

import (
	"github.com/giantswarm/versionbundle"
)

var versionBundleAzure = versionbundle.Bundle{
	Changelogs: []versionbundle.Changelog{
		{
			Component:   Name(),
			Description: "Fix bug in user values migration logic for apps.",
			Kind:        versionbundle.KindFixed,
			URLs: []string{
				"https://github.com/giantswarm/cluster-operator/pull/1030",
			},
		},
	},
	Components: []versionbundle.Component{},
	Name:       "cluster-operator",
	Provider:   "azure",
	Version:    BundleVersion(),
}
