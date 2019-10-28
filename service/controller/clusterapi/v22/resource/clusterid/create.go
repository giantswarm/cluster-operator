package clusterid

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v22/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	old, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var cr v1alpha1.Cluster
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding latest cluster")

		cl, err := r.cmaClient.ClusterV1alpha1().Clusters(old.Namespace).Get(old.Name, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		cr = *cl

		r.logger.LogCtx(ctx, "level", "debug", "message", "found latest cluster")
	}

	status := key.ClusterCommonStatus(cr)

	{
		if status.ID != "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("cluster %#q has cluster id in status", cr.Name))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		}

		if key.ClusterID(&cr) == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("cluster %#q misses cluster id in labels", cr.Name))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating cluster status")

		status.ID = key.ClusterID(&cr)
		new := r.commonClusterStatusAccessor.SetCommonClusterStatus(cr, status)

		_, err = r.cmaClient.ClusterV1alpha1().Clusters(cr.Namespace).UpdateStatus(&new)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated cluster status")

		// All further resources require cluster ID to be present in the status so
		// it makes sense to cancel whole CR reconciliation here and start from the
		// beginning.
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
	}

	return nil
}
