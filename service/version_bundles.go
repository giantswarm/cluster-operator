package service

import (
	"github.com/giantswarm/versionbundle"

	"github.com/giantswarm/cluster-operator/service/controller/aws"
	"github.com/giantswarm/cluster-operator/service/controller/azure"
	"github.com/giantswarm/cluster-operator/service/controller/kvm"
)

func NewVersionBundles(p string) []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, aws.VersionBundle())
	versionBundles = append(versionBundles, azure.VersionBundle())
	versionBundles = append(versionBundles, kvm.VersionBundle())

	return versionBundles
}
