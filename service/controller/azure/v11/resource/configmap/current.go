package configmap

import (
	"context"

	"github.com/giantswarm/errors/guest"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"

	"github.com/giantswarm/cluster-operator/pkg/v11/configmap"
	"github.com/giantswarm/cluster-operator/pkg/v11/key"
	azurekey "github.com/giantswarm/cluster-operator/service/controller/azure/v11/key"
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

	clusterConfig := configmap.ClusterConfig{
		APIDomain: apiDomain,
		ClusterID: key.ClusterID(clusterGuestConfig),
	}
	configMaps, err := r.configMap.GetCurrentState(ctx, clusterConfig)
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

	return configMaps, nil
}
