package configmap

import (
	"context"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"

	azurekey "github.com/giantswarm/cluster-operator/service/controller/azure/v22/key"
	"github.com/giantswarm/cluster-operator/service/controller/internal/configmap"
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	customObject, err := azurekey.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	configMapsToUpdate, err := toConfigMaps(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	clusterGuestConfig := azurekey.ClusterGuestConfig(customObject)
	apiDomain, err := key.APIDomain(clusterGuestConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	clusterConfig := configmap.ClusterConfig{
		APIDomain: apiDomain,
		ClusterID: key.ClusterID(clusterGuestConfig),
	}
	err = r.configMap.ApplyUpdateChange(ctx, clusterConfig, configMapsToUpdate)
	if tenant.IsAPINotAvailable(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is not available")

		// We can't continue without a successful K8s connection. Cluster
		// may not be up yet. We will retry during the next execution.
		reconciliationcanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")

		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	currentConfigMaps, err := toConfigMaps(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredConfigMaps, err := toConfigMaps(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch, err := r.configMap.NewUpdatePatch(ctx, currentConfigMaps, desiredConfigMaps)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return patch, nil
}
