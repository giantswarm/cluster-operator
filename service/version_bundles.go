package service

import (
	"github.com/giantswarm/versionbundle"

	awsv10 "github.com/giantswarm/cluster-operator/service/controller/aws/v10"
	awsv11 "github.com/giantswarm/cluster-operator/service/controller/aws/v11"
	awsv12 "github.com/giantswarm/cluster-operator/service/controller/aws/v12"
	awsv7patch1 "github.com/giantswarm/cluster-operator/service/controller/aws/v7patch1"
	awsv7patch2 "github.com/giantswarm/cluster-operator/service/controller/aws/v7patch2"
	awsv8 "github.com/giantswarm/cluster-operator/service/controller/aws/v8"
	awsv9 "github.com/giantswarm/cluster-operator/service/controller/aws/v9"
	azurev10 "github.com/giantswarm/cluster-operator/service/controller/azure/v10"
	azurev11 "github.com/giantswarm/cluster-operator/service/controller/azure/v11"
	azurev12 "github.com/giantswarm/cluster-operator/service/controller/azure/v12"
	azurev9 "github.com/giantswarm/cluster-operator/service/controller/azure/v9"
	kvmv10 "github.com/giantswarm/cluster-operator/service/controller/kvm/v10"
	kvmv11 "github.com/giantswarm/cluster-operator/service/controller/kvm/v11"
	kvmv12 "github.com/giantswarm/cluster-operator/service/controller/kvm/v12"
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

	versionBundles = append(versionBundles, awsv7patch1.VersionBundle())
	versionBundles = append(versionBundles, awsv7patch2.VersionBundle())
	versionBundles = append(versionBundles, kvmv7patch1.VersionBundle())

	versionBundles = append(versionBundles, awsv8.VersionBundle())
	versionBundles = append(versionBundles, kvmv8.VersionBundle())

	versionBundles = append(versionBundles, awsv9.VersionBundle())
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

	return versionBundles
}
