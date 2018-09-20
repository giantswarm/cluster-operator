package chartconfig

import (
	"context"

	"github.com/giantswarm/errors/guest"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"

	"github.com/giantswarm/cluster-operator/pkg/v7/chartconfig"
	"github.com/giantswarm/cluster-operator/pkg/v7/key"
	azurekey "github.com/giantswarm/cluster-operator/service/controller/azure/v7/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
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
		APIDomain: apiDomain,
		ClusterID: key.ClusterID(clusterGuestConfig),
	}
	chartConfigs, err := r.chartConfig.GetCurrentState(ctx, clusterConfig)
	if guest.IsAPINotAvailable(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is not available")

		// We can't continue without a successful K8s connection. Cluster
		// may not be up yet. We will retry during the next execution.
		reconciliationcanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")

		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	return chartConfigs, nil
}
