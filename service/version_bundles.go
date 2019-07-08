package service

import (
	"github.com/giantswarm/versionbundle"

	awsv10 "github.com/giantswarm/cluster-operator/service/controller/aws/v10"
	awsv11 "github.com/giantswarm/cluster-operator/service/controller/aws/v11"
	awsv12 "github.com/giantswarm/cluster-operator/service/controller/aws/v12"
	awsv13 "github.com/giantswarm/cluster-operator/service/controller/aws/v13"
	awsv14 "github.com/giantswarm/cluster-operator/service/controller/aws/v14"
	awsv14patch1 "github.com/giantswarm/cluster-operator/service/controller/aws/v14patch1"
	awsv15 "github.com/giantswarm/cluster-operator/service/controller/aws/v15"
	awsv16 "github.com/giantswarm/cluster-operator/service/controller/aws/v16"
	awsv17 "github.com/giantswarm/cluster-operator/service/controller/aws/v17"
	awsv18 "github.com/giantswarm/cluster-operator/service/controller/aws/v18"
	azurev10 "github.com/giantswarm/cluster-operator/service/controller/azure/v10"
	azurev11 "github.com/giantswarm/cluster-operator/service/controller/azure/v11"
	azurev12 "github.com/giantswarm/cluster-operator/service/controller/azure/v12"
	azurev13 "github.com/giantswarm/cluster-operator/service/controller/azure/v13"
	azurev14 "github.com/giantswarm/cluster-operator/service/controller/azure/v14"
	azurev14patch1 "github.com/giantswarm/cluster-operator/service/controller/azure/v14patch1"
	azurev15 "github.com/giantswarm/cluster-operator/service/controller/azure/v15"
	azurev16 "github.com/giantswarm/cluster-operator/service/controller/azure/v16"
	azurev17 "github.com/giantswarm/cluster-operator/service/controller/azure/v17"
	azurev18 "github.com/giantswarm/cluster-operator/service/controller/azure/v18"
	kvmv10 "github.com/giantswarm/cluster-operator/service/controller/kvm/v10"
	kvmv11 "github.com/giantswarm/cluster-operator/service/controller/kvm/v11"
	kvmv12 "github.com/giantswarm/cluster-operator/service/controller/kvm/v12"
	kvmv13 "github.com/giantswarm/cluster-operator/service/controller/kvm/v13"
	kvmv14 "github.com/giantswarm/cluster-operator/service/controller/kvm/v14"
	kvmv14patch1 "github.com/giantswarm/cluster-operator/service/controller/kvm/v14patch1"
	kvmv15 "github.com/giantswarm/cluster-operator/service/controller/kvm/v15"
	kvmv16 "github.com/giantswarm/cluster-operator/service/controller/kvm/v16"
	kvmv17 "github.com/giantswarm/cluster-operator/service/controller/kvm/v17"
	kvmv18 "github.com/giantswarm/cluster-operator/service/controller/kvm/v18"
)

func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

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

	versionBundles = append(versionBundles, awsv14.VersionBundle())
	versionBundles = append(versionBundles, azurev14.VersionBundle())
	versionBundles = append(versionBundles, kvmv14.VersionBundle())

	versionBundles = append(versionBundles, awsv14patch1.VersionBundle())
	versionBundles = append(versionBundles, azurev14patch1.VersionBundle())
	versionBundles = append(versionBundles, kvmv14patch1.VersionBundle())

	versionBundles = append(versionBundles, awsv15.VersionBundle())
	versionBundles = append(versionBundles, azurev15.VersionBundle())
	versionBundles = append(versionBundles, kvmv15.VersionBundle())

	versionBundles = append(versionBundles, awsv16.VersionBundle())
	versionBundles = append(versionBundles, azurev16.VersionBundle())
	versionBundles = append(versionBundles, kvmv16.VersionBundle())

	versionBundles = append(versionBundles, awsv17.VersionBundle())
	versionBundles = append(versionBundles, azurev17.VersionBundle())
	versionBundles = append(versionBundles, kvmv17.VersionBundle())

	versionBundles = append(versionBundles, awsv18.VersionBundle())
	versionBundles = append(versionBundles, azurev18.VersionBundle())
	versionBundles = append(versionBundles, kvmv18.VersionBundle())

	return versionBundles
}
