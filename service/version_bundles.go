package service

import (
	"github.com/giantswarm/versionbundle"

	awsv1 "github.com/giantswarm/cluster-operator/service/controller/aws/v1"
	awsv2 "github.com/giantswarm/cluster-operator/service/controller/aws/v2"
	awsv3 "github.com/giantswarm/cluster-operator/service/controller/aws/v3"
	awsv4 "github.com/giantswarm/cluster-operator/service/controller/aws/v4"
	awsv5 "github.com/giantswarm/cluster-operator/service/controller/aws/v5"
	awsv6 "github.com/giantswarm/cluster-operator/service/controller/aws/v6"
	azurev1 "github.com/giantswarm/cluster-operator/service/controller/azure/v1"
	azurev2 "github.com/giantswarm/cluster-operator/service/controller/azure/v2"
	azurev3 "github.com/giantswarm/cluster-operator/service/controller/azure/v3"
	azurev4 "github.com/giantswarm/cluster-operator/service/controller/azure/v4"
	azurev5 "github.com/giantswarm/cluster-operator/service/controller/azure/v5"
	azurev6 "github.com/giantswarm/cluster-operator/service/controller/azure/v6"
	kvmv1 "github.com/giantswarm/cluster-operator/service/controller/kvm/v1"
	kvmv2 "github.com/giantswarm/cluster-operator/service/controller/kvm/v2"
	kvmv3 "github.com/giantswarm/cluster-operator/service/controller/kvm/v3"
	kvmv4 "github.com/giantswarm/cluster-operator/service/controller/kvm/v4"
	kvmv5 "github.com/giantswarm/cluster-operator/service/controller/kvm/v5"
	kvmv6 "github.com/giantswarm/cluster-operator/service/controller/kvm/v6"
)

func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, awsv1.VersionBundle())
	versionBundles = append(versionBundles, azurev1.VersionBundle())
	versionBundles = append(versionBundles, kvmv1.VersionBundle())

	versionBundles = append(versionBundles, awsv2.VersionBundle())
	versionBundles = append(versionBundles, azurev2.VersionBundle())
	versionBundles = append(versionBundles, kvmv2.VersionBundle())

	versionBundles = append(versionBundles, awsv3.VersionBundle())
	versionBundles = append(versionBundles, azurev3.VersionBundle())
	versionBundles = append(versionBundles, kvmv3.VersionBundle())

	versionBundles = append(versionBundles, awsv4.VersionBundle())
	versionBundles = append(versionBundles, azurev4.VersionBundle())
	versionBundles = append(versionBundles, kvmv4.VersionBundle())

	versionBundles = append(versionBundles, awsv5.VersionBundle())
	versionBundles = append(versionBundles, azurev5.VersionBundle())
	versionBundles = append(versionBundles, kvmv5.VersionBundle())

	versionBundles = append(versionBundles, awsv6.VersionBundle())
	versionBundles = append(versionBundles, azurev6.VersionBundle())
	versionBundles = append(versionBundles, kvmv6.VersionBundle())

	return versionBundles
}
