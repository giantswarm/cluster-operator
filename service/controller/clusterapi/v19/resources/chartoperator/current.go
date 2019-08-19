package chartoperator

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/key"
)

// GetCurrentState gets the state of the chart in the guest cluster.
func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	{
		if cc.Client.TenantCluster.Helm == nil {
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster clients not available")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil, nil
		}

		// Tenant chart-operator is not deleted so cancel the resource. The operator
		// will be deleted when the tenant cluster resources are deleted.
		if key.IsDeleted(&cr) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not deleting chart-operator %#q in tenant cluster %#q", namespace, key.ClusterID(&cr)))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil, nil
		}
	}

	var releaseContent *helmclient.ReleaseContent
	var releaseHistory *helmclient.ReleaseHistory
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding chart-operator release %#q in tenant cluster %#q", release, key.ClusterID(&cr)))

		releaseContent, err = cc.Client.TenantCluster.Helm.GetReleaseContent(ctx, release)
		if helmclient.IsReleaseNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find chart-operator release %#q in tenant cluster %#q", release, key.ClusterID(&cr)))
			return nil, nil

		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		releaseHistory, err = cc.Client.TenantCluster.Helm.GetReleaseHistory(ctx, release)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found chart-operator release %#q in tenant cluster %#q", release, key.ClusterID(&cr)))
	}

	var chartState *ResourceState
	{
		bytes, err := json.Marshal(releaseContent.Values)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		chartValues := &Values{}
		err = json.Unmarshal(bytes, chartValues)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		chartState = &ResourceState{
			ChartName:      chart,
			ChartValues:    *chartValues,
			ReleaseName:    release,
			ReleaseStatus:  releaseContent.Status,
			ReleaseVersion: releaseHistory.Version,
		}
	}

	return chartState, nil
}
