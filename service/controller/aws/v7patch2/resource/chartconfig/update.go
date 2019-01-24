package chartconfig

import (
	"context"

	"github.com/giantswarm/errors/guest"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"

	"github.com/giantswarm/cluster-operator/pkg/v7patch2/chartconfig"
	"github.com/giantswarm/cluster-operator/pkg/v7patch2/key"
	awskey "github.com/giantswarm/cluster-operator/service/controller/aws/v7patch2/key"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	customObject, err := awskey.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	chartConfigsToUpdate, err := toChartConfigs(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	clusterGuestConfig := awskey.ClusterGuestConfig(customObject)
	apiDomain, err := key.APIDomain(clusterGuestConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	clusterConfig := chartconfig.ClusterConfig{
		APIDomain:    apiDomain,
		ClusterID:    key.ClusterID(clusterGuestConfig),
		Organization: key.ClusterOrganization(clusterGuestConfig),
	}

	err = r.chartConfig.ApplyUpdateChange(ctx, clusterConfig, chartConfigsToUpdate)
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

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	currentChartConfigs, err := toChartConfigs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredChartConfigs, err := toChartConfigs(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch, err := r.chartConfig.NewUpdatePatch(ctx, currentChartConfigs, desiredChartConfigs)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return patch, nil
}
