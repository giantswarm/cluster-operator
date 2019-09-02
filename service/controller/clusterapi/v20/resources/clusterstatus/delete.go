package clusterstatus

import (
	"context"
	"fmt"
	"reflect"

	"github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v20/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	updatedClusterStatus := r.computeDeleteClusterConditions(ctx, r.accessor.GetCommonClusterStatus(cr))

	if !reflect.DeepEqual(r.accessor.GetCommonClusterStatus(cr), updatedClusterStatus) {
		{
			r.logger.LogCtx(ctx, "level", "debug", "message", "updating cluster status")

			cr = r.accessor.SetCommonClusterStatus(cr, updatedClusterStatus)
			_, err := r.cmaClient.ClusterV1alpha1().Clusters(cr.Namespace).UpdateStatus(&cr)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", "updated cluster status")
		}

		{
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
			reconciliationcanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "keeping finalizers")
			finalizerskeptcontext.SetKept(ctx)
		}

		return nil
	}

	return nil
}

func (r *Resource) computeDeleteClusterConditions(ctx context.Context, clusterStatus v1alpha1.CommonClusterStatus) v1alpha1.CommonClusterStatus {
	// On Deletion we always add the deleting status condition.
	// We skip adding the condition if it's already set.
	{
		notDeleting := !clusterStatus.HasDeletingCondition()
		if notDeleting {
			clusterStatus.Conditions = clusterStatus.WithDeletingCondition()
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", v1alpha1.ClusterStatusConditionDeleting))
		}
	}

	return clusterStatus
}
