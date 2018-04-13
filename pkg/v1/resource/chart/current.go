package chart

import (
	"context"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework/context/resourcecanceledcontext"

	"github.com/giantswarm/cluster-operator/pkg/v1/guestcluster"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	clusterGuestConfig, err := r.toClusterGuestConfigFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "looking for chart-operator chart in the guest cluster")

	clusterConfig, err := prepareClusterConfig(r.baseClusterConfig, clusterGuestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	guestHelmClient, err := r.guest.NewHelmClient(ctx, clusterConfig.ClusterID, clusterConfig.Domain.API)
	if guestcluster.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the cluster-operator api cert in the Kubernetes API")

		// We can't continue without the cluster-operator cert. We will retry
		// during the next execution.
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")

		return nil, nil

	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	releaseContent, err := guestHelmClient.GetReleaseContent(chartOperatorRelease)
	if helmclient.IsReleaseNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the chart-operator chart in the guest cluster")
		return nil, nil

	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	releaseHistory, err := guestHelmClient.GetReleaseHistory(chartOperatorRelease)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	chartState := &ResourceState{
		ChartName:      chartOperatorChart,
		ReleaseName:    chartOperatorRelease,
		ReleaseStatus:  releaseContent.Status,
		ReleaseVersion: releaseHistory.Version,
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "found chart-operator chart in the guest cluster")

	return chartState, nil
}
