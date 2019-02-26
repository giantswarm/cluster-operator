package chartconfig

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/cluster-operator/pkg/v12/chartconfig"
	"github.com/giantswarm/cluster-operator/pkg/v12/key"
	azurekey "github.com/giantswarm/cluster-operator/service/controller/azure/v12/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := azurekey.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterGuestConfig := azurekey.ClusterGuestConfig(customObject)
	apiDomain, err := key.APIDomain(clusterGuestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterConfig := chartconfig.ClusterConfig{
		APIDomain:    apiDomain,
		ClusterID:    key.ClusterID(clusterGuestConfig),
		Organization: key.ClusterOrganization(clusterGuestConfig),
	}

	desiredConfigMaps, err := r.chartConfig.GetDesiredState(ctx, clusterConfig, azurekey.ChartSpecs())
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return desiredConfigMaps, nil
}
