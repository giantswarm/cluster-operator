package v4

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "cluster-operator",
				Description: "Added chartconfig resource.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "cluster-operator",
				Description: "Added external-dns chartconfig resource.",
				Kind:        versionbundle.KindAdded,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "nginx-ingress-controller",
				Version: "0.12.0",
			},
			{
				Name:    "external-dns",
				Version: "0.5.2",
			},
		},
		Name:     "cluster-operator",
		Provider: "azure",
		Version:  "0.4.0",
	}
}
