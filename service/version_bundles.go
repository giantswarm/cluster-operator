package service

import (
	"github.com/giantswarm/versionbundle"

	awsv14 "github.com/giantswarm/cluster-operator/service/controller/aws/v14"
	awsv14patch1 "github.com/giantswarm/cluster-operator/service/controller/aws/v14patch1"
	awsv15 "github.com/giantswarm/cluster-operator/service/controller/aws/v15"
	awsv16 "github.com/giantswarm/cluster-operator/service/controller/aws/v16"
	awsv17 "github.com/giantswarm/cluster-operator/service/controller/aws/v17"
	awsv18 "github.com/giantswarm/cluster-operator/service/controller/aws/v18"
	awsv19 "github.com/giantswarm/cluster-operator/service/controller/aws/v19"
	awsv20 "github.com/giantswarm/cluster-operator/service/controller/aws/v20"
	awsv21 "github.com/giantswarm/cluster-operator/service/controller/aws/v21"
	azurev14 "github.com/giantswarm/cluster-operator/service/controller/azure/v14"
	azurev14patch1 "github.com/giantswarm/cluster-operator/service/controller/azure/v14patch1"
	azurev15 "github.com/giantswarm/cluster-operator/service/controller/azure/v15"
	azurev16 "github.com/giantswarm/cluster-operator/service/controller/azure/v16"
	azurev17 "github.com/giantswarm/cluster-operator/service/controller/azure/v17"
	azurev18 "github.com/giantswarm/cluster-operator/service/controller/azure/v18"
	azurev19 "github.com/giantswarm/cluster-operator/service/controller/azure/v19"
	azurev20 "github.com/giantswarm/cluster-operator/service/controller/azure/v20"
	azurev21 "github.com/giantswarm/cluster-operator/service/controller/azure/v21"
	clusterapiv21 "github.com/giantswarm/cluster-operator/service/controller/clusterapi/v21"
	kvmv14patch1 "github.com/giantswarm/cluster-operator/service/controller/kvm/v14patch1"
	kvmv15 "github.com/giantswarm/cluster-operator/service/controller/kvm/v15"
	kvmv16 "github.com/giantswarm/cluster-operator/service/controller/kvm/v16"
	kvmv17 "github.com/giantswarm/cluster-operator/service/controller/kvm/v17"
	kvmv18 "github.com/giantswarm/cluster-operator/service/controller/kvm/v18"
	kvmv19 "github.com/giantswarm/cluster-operator/service/controller/kvm/v19"
	kvmv20 "github.com/giantswarm/cluster-operator/service/controller/kvm/v20"
	kvmv21 "github.com/giantswarm/cluster-operator/service/controller/kvm/v21"
)

func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, awsv14.VersionBundle())
	versionBundles = append(versionBundles, azurev14.VersionBundle())

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

	versionBundles = append(versionBundles, awsv19.VersionBundle())
	versionBundles = append(versionBundles, azurev19.VersionBundle())
	versionBundles = append(versionBundles, kvmv19.VersionBundle())

	versionBundles = append(versionBundles, awsv20.VersionBundle())
	versionBundles = append(versionBundles, azurev20.VersionBundle())
	versionBundles = append(versionBundles, kvmv20.VersionBundle())

	versionBundles = append(versionBundles, awsv21.VersionBundle())
	versionBundles = append(versionBundles, azurev21.VersionBundle())
	versionBundles = append(versionBundles, clusterapiv21.VersionBundle())
	versionBundles = append(versionBundles, kvmv21.VersionBundle())

	return versionBundles
}
