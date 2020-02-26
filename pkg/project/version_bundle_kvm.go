package project

import (
	"github.com/giantswarm/versionbundle"
)

var versionBundleKVM = versionbundle.Bundle{
	Changelogs: []versionbundle.Changelog{
		{
			Component:   "cluster-operator",
			Description: "User app override rules applied",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/cluster-operator/pull/930",
			},
		},
	},
	Components: []versionbundle.Component{},
	Name:       "cluster-operator",
	Provider:   "kvm",
	Version:    BundleVersion(),
}
