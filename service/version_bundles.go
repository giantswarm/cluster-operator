package service

import (
	"github.com/giantswarm/versionbundle"

	awsv1 "github.com/giantswarm/cluster-operator/service/awsclusterconfig/v1"
	azurev1 "github.com/giantswarm/cluster-operator/service/azureclusterconfig/v1"
	kvmv1 "github.com/giantswarm/cluster-operator/service/kvmclusterconfig/v1"

	awsv2 "github.com/giantswarm/cluster-operator/service/awsclusterconfig/v2"
	azurev2 "github.com/giantswarm/cluster-operator/service/azureclusterconfig/v2"
	kvmv2 "github.com/giantswarm/cluster-operator/service/kvmclusterconfig/v2"
)

func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, awsv1.VersionBundle())
	versionBundles = append(versionBundles, azurev1.VersionBundle())
	versionBundles = append(versionBundles, kvmv1.VersionBundle())

	versionBundles = append(versionBundles, awsv2.VersionBundle())
	versionBundles = append(versionBundles, azurev2.VersionBundle())
	versionBundles = append(versionBundles, kvmv2.VersionBundle())

	return versionBundles
}
