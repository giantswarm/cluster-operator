package service

import (
	"github.com/giantswarm/versionbundle"

	awsv1 "github.com/giantswarm/cluster-operator/service/awsclusterconfig/v1"
	kvmv1 "github.com/giantswarm/cluster-operator/service/kvmclusterconfig/v1"
)

// NewVersionBundles returns available version bundles based on given provider.
// Provider parameter must be validated to be one of ['aws', 'kvm'].
func NewVersionBundles(provider string) []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	switch provider {
	case "aws":
		versionBundles = append(versionBundles, awsv1.VersionBundle())
	case "kvm":
		versionBundles = append(versionBundles, kvmv1.VersionBundle())
	}

	return versionBundles
}
