package service

import (
	"github.com/giantswarm/versionbundle"

	awsv1 "github.com/giantswarm/cluster-operator/service/awsclusterconfig/v1"
	azurev1 "github.com/giantswarm/cluster-operator/service/azureclusterconfig/v1"
	kvmv1 "github.com/giantswarm/cluster-operator/service/kvmclusterconfig/v1"
)

func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, awsv1.VersionBundle())
	versionBundles = append(versionBundles, azurev1.VersionBundle())
	versionBundles = append(versionBundles, kvmv1.VersionBundle())

	return versionBundles
}
