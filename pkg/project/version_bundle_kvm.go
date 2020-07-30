package project

import (
	"github.com/giantswarm/versionbundle"
)

var versionBundleKVM = versionbundle.Bundle{
	Changelogs: []versionbundle.Changelog{
		{
			Component:   Name(),
			Description: "Make NGINX optional on KVM, by ignoring existing NGINX IC App CRs which were managed by cluster-operator.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/cluster-operator/pull/1170",
			},
		},
	},
	Components: []versionbundle.Component{},
	Name:       "cluster-operator",
	Provider:   "kvm",
	Version:    BundleVersion(),
}
