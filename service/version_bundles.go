package service

import (
	"github.com/giantswarm/versionbundle"

	awsv10 "github.com/giantswarm/cluster-operator/service/controller/aws/v10"
	awsv11 "github.com/giantswarm/cluster-operator/service/controller/aws/v11"
	awsv12 "github.com/giantswarm/cluster-operator/service/controller/aws/v12"
	awsv13 "github.com/giantswarm/cluster-operator/service/controller/aws/v13"
	awsv14 "github.com/giantswarm/cluster-operator/service/controller/aws/v14"
	awsv15 "github.com/giantswarm/cluster-operator/service/controller/aws/v15"
	azurev10 "github.com/giantswarm/cluster-operator/service/controller/azure/v10"
	azurev11 "github.com/giantswarm/cluster-operator/service/controller/azure/v11"
	azurev12 "github.com/giantswarm/cluster-operator/service/controller/azure/v12"
	azurev13 "github.com/giantswarm/cluster-operator/service/controller/azure/v13"
	azurev14 "github.com/giantswarm/cluster-operator/service/controller/azure/v14"
	azurev15 "github.com/giantswarm/cluster-operator/service/controller/azure/v15"
	azurev9 "github.com/giantswarm/cluster-operator/service/controller/azure/v9"
	kvmv10 "github.com/giantswarm/cluster-operator/service/controller/kvm/v10"
	kvmv11 "github.com/giantswarm/cluster-operator/service/controller/kvm/v11"
	kvmv12 "github.com/giantswarm/cluster-operator/service/controller/kvm/v12"
	kvmv13 "github.com/giantswarm/cluster-operator/service/controller/kvm/v13"
	kvmv13patch1 "github.com/giantswarm/cluster-operator/service/controller/kvm/v13patch1"
	kvmv14 "github.com/giantswarm/cluster-operator/service/controller/kvm/v14"
	kvmv15 "github.com/giantswarm/cluster-operator/service/controller/kvm/v15"
	kvmv6 "github.com/giantswarm/cluster-operator/service/controller/kvm/v6"
	kvmv6patch1 "github.com/giantswarm/cluster-operator/service/controller/kvm/v6patch1"
	kvmv7 "github.com/giantswarm/cluster-operator/service/controller/kvm/v7"
	kvmv7patch1 "github.com/giantswarm/cluster-operator/service/controller/kvm/v7patch1"
	kvmv8 "github.com/giantswarm/cluster-operator/service/controller/kvm/v8"
	kvmv9 "github.com/giantswarm/cluster-operator/service/controller/kvm/v9"
)

func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, kvmv6.VersionBundle())
	versionBundles = append(versionBundles, kvmv6patch1.VersionBundle())

	versionBundles = append(versionBundles, kvmv7.VersionBundle())

	versionBundles = append(versionBundles, kvmv7patch1.VersionBundle())

	versionBundles = append(versionBundles, kvmv8.VersionBundle())

	versionBundles = append(versionBundles, azurev9.VersionBundle())
	versionBundles = append(versionBundles, kvmv9.VersionBundle())

	versionBundles = append(versionBundles, awsv10.VersionBundle())
	versionBundles = append(versionBundles, azurev10.VersionBundle())
	versionBundles = append(versionBundles, kvmv10.VersionBundle())

	versionBundles = append(versionBundles, awsv11.VersionBundle())
	versionBundles = append(versionBundles, azurev11.VersionBundle())
	versionBundles = append(versionBundles, kvmv11.VersionBundle())

	versionBundles = append(versionBundles, awsv12.VersionBundle())
	versionBundles = append(versionBundles, azurev12.VersionBundle())
	versionBundles = append(versionBundles, kvmv12.VersionBundle())

	versionBundles = append(versionBundles, awsv13.VersionBundle())
	versionBundles = append(versionBundles, azurev13.VersionBundle())
	versionBundles = append(versionBundles, kvmv13.VersionBundle())
	versionBundles = append(versionBundles, kvmv13patch1.VersionBundle())

	versionBundles = append(versionBundles, awsv14.VersionBundle())
	versionBundles = append(versionBundles, azurev14.VersionBundle())
	versionBundles = append(versionBundles, kvmv14.VersionBundle())

	versionBundles = append(versionBundles, awsv15.VersionBundle())
	versionBundles = append(versionBundles, azurev15.VersionBundle())
	versionBundles = append(versionBundles, kvmv15.VersionBundle())

	return versionBundles
}
