package v5

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "cluster-operator",
				Description: "Added configmap resource for managing chart configuration in guest clusters.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "cluster-operator",
				Description: "Configured the Ingress Controller chart with the number of worker nodes.",
				Kind:        versionbundle.KindChanged,
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
		Version:  "0.5.0",
	}
}
