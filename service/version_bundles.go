package service

import (
	"github.com/giantswarm/versionbundle"

	awsv22 "github.com/giantswarm/cluster-operator/service/controller/aws"
	azurev22 "github.com/giantswarm/cluster-operator/service/controller/azure"
	"github.com/giantswarm/cluster-operator/service/controller/kvm"
)

func NewVersionBundles(p string) []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, awsv22.VersionBundle())
	versionBundles = append(versionBundles, azurev22.VersionBundle())
	versionBundles = append(versionBundles, kvm.VersionBundle())

	return versionBundles
}
