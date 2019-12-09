package clusterstatus

import (
	"context"
	"fmt"
	"reflect"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"

	"github.com/giantswarm/cluster-operator/service/controller/key"
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

func (r *Resource) computeDeleteClusterConditions(ctx context.Context, clusterStatus infrastructurev1alpha2.CommonClusterStatus) infrastructurev1alpha2.CommonClusterStatus {
	// On Deletion we always add the deleting status condition.
	// We skip adding the condition if it's already set.
	{
		notDeleting := !clusterStatus.HasDeletingCondition()
		if notDeleting {
			clusterStatus.Conditions = clusterStatus.WithDeletingCondition()
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", infrastructurev1alpha2.ClusterStatusConditionDeleting))
		}
	}

	return clusterStatus
}
