package service

import (
	"github.com/giantswarm/versionbundle"

	clusterapiv22 "github.com/giantswarm/cluster-operator/service/controller/clusterapi/v22"
)

func NewVersionBundles(p string) []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, clusterapiv22.VersionBundle(p))

	return versionBundles
}
