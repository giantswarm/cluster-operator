package project

import (
	"github.com/giantswarm/versionbundle"
)

var versionBundleKVM = versionbundle.Bundle{
	Changelogs: []versionbundle.Changelog{
		{
			Component:   "cluster-operator",
			Description: "Support small cluster profile detection.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/cluster-operator/pull/960",
			},
		},
	},
	Components: []versionbundle.Component{},
	Name:       "cluster-operator",
	Provider:   "kvm",
	Version:    BundleVersion(),
}
