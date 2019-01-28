package chart

import (
	"context"

	"github.com/giantswarm/errors/guest"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	"github.com/giantswarm/tenantcluster"
)

// GetCurrentState gets the state of the chart in the guest cluster.
func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	var err error

	var tenantHelmClient helmclient.Interface
	{
		tenantHelmClient, err = r.getTenantHelmClient(ctx, obj)
		if tenantcluster.IsTimeout(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "timeout fetching certificates")

			// We can't continue without a Helm client. We will retry during the next
			// execution.
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			reconciliationcanceledcontext.SetCanceled(ctx)

			return nil, nil
		} else if guest.IsAPINotAvailable(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "guest cluster is not available")

			// We can't continue without a successful K8s connection. Cluster may not
			// be up yet. We will retry during the next execution.
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			reconciliationcanceledcontext.SetCanceled(ctx)

			return nil, nil
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var chartState *ResourceState
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the chart-operator chart in the guest cluster")

		releaseContent, err := tenantHelmClient.GetReleaseContent(ctx, chartOperatorRelease)
		if helmclient.IsReleaseNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the chart-operator chart in the guest cluster")

			return nil, nil
		} else if guest.IsAPINotAvailable(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "guest cluster is not available")

			// We can't continue without a successful K8s connection. Cluster may not
			// be up yet. We will retry during the next execution.
			reconciliationcanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil, nil
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		releaseHistory, err := tenantHelmClient.GetReleaseHistory(ctx, chartOperatorRelease)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		chartState = &ResourceState{
			ChartName:      chartOperatorChart,
			ReleaseName:    chartOperatorRelease,
			ReleaseStatus:  releaseContent.Status,
			ReleaseVersion: releaseHistory.Version,
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found the chart-operator chart in the guest cluster")
	}

	return chartState, nil
}
