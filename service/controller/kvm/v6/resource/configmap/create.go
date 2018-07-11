package configmap

import (
	"context"

	"github.com/giantswarm/errors/guest"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"

	"github.com/giantswarm/cluster-operator/pkg/v6/configmap"
	"github.com/giantswarm/cluster-operator/pkg/v6/key"
	kvmkey "github.com/giantswarm/cluster-operator/service/controller/kvm/v6/key"
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
	guestAPIDomain, err := key.APIDomain(clusterGuestConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	configMapConfig := configmap.ConfigMapConfig{
		ClusterID:      key.ClusterID(clusterGuestConfig),
		GuestAPIDomain: guestAPIDomain,
	}
	err = r.configMap.ApplyCreateChange(ctx, configMapConfig, configMapsToCreate)
	if guest.IsAPINotAvailable(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "guest cluster is not available")

		// We can't continue without a successful K8s connection. Cluster
		// may not be up yet. We will retry during the next execution.
		reconciliationcanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")

		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
