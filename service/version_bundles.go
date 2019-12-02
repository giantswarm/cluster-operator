package service

import (
	"github.com/giantswarm/versionbundle"

	"github.com/giantswarm/cluster-operator/pkg/project"
)

func NewVersionBundles(p string) []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, project.VersionBundle(p))

	return versionBundles
}
