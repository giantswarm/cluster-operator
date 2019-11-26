package project

import (
	"github.com/giantswarm/versionbundle"
)

func NewVersionBundles() []versionbundle.Bundle {
	return []versionbundle.Bundle{
		versionBundleAWS,
		versionBundleAzure,
		versionBundleKVM,
	}
}
