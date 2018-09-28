package configmap

import (
	"context"

	"github.com/giantswarm/errors/guest"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"

	"github.com/giantswarm/cluster-operator/pkg/v7/configmap"
	"github.com/giantswarm/cluster-operator/pkg/v7/key"
	kvmkey "github.com/giantswarm/cluster-operator/service/controller/kvm/v7/key"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	customObject, err := kvmkey.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	configMapsToDelete, err := toConfigMaps(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	clusterGuestConfig := kvmkey.ClusterGuestConfig(customObject)
	apiDomain, err := key.APIDomain(clusterGuestConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	clusterConfig := configmap.ClusterConfig{
		APIDomain: apiDomain,
		ClusterID: key.ClusterID(clusterGuestConfig),
	}
	err = r.configMap.ApplyDeleteChange(ctx, clusterConfig, configMapsToDelete)
	if guest.IsAPINotAvailable(err) {
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

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	currentConfigMaps, err := toConfigMaps(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredConfigMaps, err := toConfigMaps(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch, err := r.configMap.NewDeletePatch(ctx, currentConfigMaps, desiredConfigMaps)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return patch, nil
}
