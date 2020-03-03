package project

import (
	"github.com/giantswarm/versionbundle"
)

var versionBundleAzure = versionbundle.Bundle{
	Changelogs: []versionbundle.Changelog{
		{
			Component:   "cluster-operator",
			Description: "Stop setting IC replicas count",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/cluster-operator/pull/950",
			},
		},
	},
	Components: []versionbundle.Component{},
	Name:       "cluster-operator",
	Provider:   "azure",
	Version:    BundleVersion(),
}
