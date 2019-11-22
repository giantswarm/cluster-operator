package chartconfig

import (
	"context"

	"github.com/giantswarm/microerror"

	awskey "github.com/giantswarm/cluster-operator/service/controller/aws/v22/key"
	"github.com/giantswarm/cluster-operator/service/controller/internal/chartconfig"
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := awskey.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterGuestConfig := awskey.ClusterGuestConfig(customObject)
	apiDomain, err := key.APIDomain(clusterGuestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterConfig := chartconfig.ClusterConfig{
		APIDomain:    apiDomain,
		ClusterID:    key.ClusterID(clusterGuestConfig),
		Organization: key.ClusterOrganization(clusterGuestConfig),
	}

	desiredConfigMaps, err := r.chartConfig.GetDesiredState(ctx, clusterConfig, awskey.ChartSpecs())
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return desiredConfigMaps, nil
}
