package chartconfig

import (
	"context"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"

	"github.com/giantswarm/cluster-operator/pkg/v18/chartconfig"
	"github.com/giantswarm/cluster-operator/pkg/v18/key"
	azurekey "github.com/giantswarm/cluster-operator/service/controller/azure/v18/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	customObject, err := azurekey.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	chartConfigsToCreate, err := toChartConfigs(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	clusterGuestConfig := azurekey.ClusterGuestConfig(customObject)
	apiDomain, err := key.APIDomain(clusterGuestConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	clusterConfig := chartconfig.ClusterConfig{
		APIDomain:    apiDomain,
		ClusterID:    key.ClusterID(clusterGuestConfig),
		Organization: key.ClusterOrganization(clusterGuestConfig),
	}

	err = r.chartConfig.ApplyCreateChange(ctx, clusterConfig, chartConfigsToCreate)
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
