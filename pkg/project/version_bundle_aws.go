package project

import (
	"github.com/giantswarm/versionbundle"
)

var versionBundleAWS = versionbundle.Bundle{
	Changelogs: []versionbundle.Changelog{
		{
			Component:   "cluster-operator",
			Description: "Support xxs, xs, and small cluster profile detection; xxs is new xs; xs and small profile rules are based on determined worker max CPU cores which is currently limited to / known for kvm provider only.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/cluster-operator/pull/963",
			},
		},
	},
	Components: []versionbundle.Component{},
	Name:       "cluster-operator",
	Provider:   "aws",
	Version:    BundleVersion(),
}
