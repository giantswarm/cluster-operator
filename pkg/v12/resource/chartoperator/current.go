package chartoperator

import (
	"context"

	"github.com/giantswarm/errors/guest"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"github.com/giantswarm/tenantcluster"

	"github.com/giantswarm/cluster-operator/pkg/v11/key"
)

// GetCurrentState gets the state of the chart in the guest cluster.
func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	var err error

	objectMeta, err := r.toClusterObjectMetaFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Tenant chart-operator is not deleted so cancel the resource. The operator
	// will be deleted when the tenant cluster resources are deleted.
	if key.IsDeleted(objectMeta) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "redirecting chartoperator deletion to provider operators")
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil, nil
	}

	var tenantHelmClient helmclient.Interface
	{
		tenantHelmClient, err = r.getTenantHelmClient(ctx, obj)
		if tenantcluster.IsTimeout(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "timeout fetching certificates")

			// A timeout error here means that the cluster-operator certificate for
			// the current guest cluster was not found. We can't continue without a
			// Helm client. We will retry during the next execution, when the
			// certificate might be available.
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)

			return nil, nil
		} else if helmclient.IsTillerInstallationFailed(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "Tiller installation failed")

			// Tiller installation can fail during guest cluster setup. We will retry
			// on next reconciliation loop.
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)

			return nil, nil
		} else if guest.IsAPINotAvailable(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "guest API not available")

			// We should not hammer guest API if it is not available, the guest
			// cluster might be initializing. We will retry on next reconciliation
			// loop.
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)

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
