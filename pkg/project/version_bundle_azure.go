package project

import (
	"github.com/giantswarm/versionbundle"
)

var versionBundleAzure = versionbundle.Bundle{
	Changelogs: []versionbundle.Changelog{
		{
			Component:   Name(),
			Description: "TODO",
			Kind:        versionbundle.KindChanged,
			URLs:        []string{},
		},
	},
	Components: []versionbundle.Component{},
	Name:       "cluster-operator",
	Provider:   "azure",
	Version:    BundleVersion(),
}
