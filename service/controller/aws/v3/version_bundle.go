package v3

import (
	"time"

	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "cluster-operator",
				Description: "Added giantswarm namespace.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "cluster-operator",
				Description: "Add node-exporter and kube-state-metrics chartconfigs.",
				Kind:        versionbundle.KindAdded,
			},
		},
		Components:   []versionbundle.Component{},
		Dependencies: []versionbundle.Dependency{},
		Deprecated:   false,
		Name:         "cluster-operator",
		Provider:     "aws",
		Time:         time.Date(2018, time.April, 26, 12, 00, 0, 0, time.UTC),
		Version:      "0.3.0",
		WIP:          true,
	}
}
