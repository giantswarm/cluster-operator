package clusterid

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v17/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring cluster status has cluster ID")

	status := r.commonClusterStatusAccessor.GetCommonClusterStatus(cr)

	if status.ID != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "ensured cluster status has cluster ID")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	}

	clusterID := key.ClusterID(&cr)

	if clusterID == "" {
		return microerror.Maskf(executionFailedError, "cluster ID missing from CR")
	}

	status.ID = clusterID

	updatedCR := r.commonClusterStatusAccessor.SetCommonClusterStatus(cr, status)

	r.logger.LogCtx(ctx, "level", "debug", "message", "updating cluster status")

	_, err = r.cmaClient.ClusterV1alpha1().Clusters(cr.Namespace).Update(&updatedCR)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "updated cluster status")
	r.logger.LogCtx(ctx, "level", "debug", "message", "ensured cluster status has cluster ID")
	r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")

	reconciliationcanceledcontext.SetCanceled(ctx)

	return nil
}
