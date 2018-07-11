package configmap

import (
	"context"

	"github.com/giantswarm/errors/guest"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"

	"github.com/giantswarm/cluster-operator/pkg/v6/configmap"
	"github.com/giantswarm/cluster-operator/pkg/v6/key"
	awskey "github.com/giantswarm/cluster-operator/service/controller/aws/v6/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := awskey.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterGuestConfig := awskey.ClusterGuestConfig(customObject)
	guestAPIDomain, err := key.APIDomain(clusterGuestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	configMapConfig := configmap.ConfigMapConfig{
		ClusterID:      key.ClusterID(clusterGuestConfig),
		GuestAPIDomain: guestAPIDomain,
	}
	configMaps, err := r.configMap.GetCurrentState(ctx, configMapConfig)
	if guest.IsAPINotAvailable(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "guest cluster is not available")

		// We can't continue without a successful K8s connection. Cluster
		// may not be up yet. We will retry during the next execution.
		reconciliationcanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")

		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	return configMaps, nil
}
