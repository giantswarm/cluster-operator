package configmap

import (
	"context"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"

	"github.com/giantswarm/cluster-operator/pkg/v23/configmap"
	"github.com/giantswarm/cluster-operator/pkg/v23/key"
	kvmkey "github.com/giantswarm/cluster-operator/service/controller/kvm/v23/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	customObject, err := kvmkey.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	configMapsToCreate, err := toConfigMaps(createChange)
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
	err = r.configMap.ApplyCreateChange(ctx, clusterConfig, configMapsToCreate)
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
