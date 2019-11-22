package service

import (
	"github.com/giantswarm/versionbundle"

	awsv22 "github.com/giantswarm/cluster-operator/service/controller/aws/v22"
	azurev22 "github.com/giantswarm/cluster-operator/service/controller/azure/v22"
	kvmv22 "github.com/giantswarm/cluster-operator/service/controller/kvm/v22"
)

func NewVersionBundles(p string) []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, awsv22.VersionBundle())
	versionBundles = append(versionBundles, azurev22.VersionBundle())
	versionBundles = append(versionBundles, kvmv22.VersionBundle())

	return versionBundles
}
