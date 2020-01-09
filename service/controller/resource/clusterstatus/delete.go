package clusterstatus

import (
	"context"
	"fmt"
	"reflect"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/cluster-operator/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr := r.newCommonClusterObjectFunc()
	{
		cl, err := key.ToCluster(obj)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding latest infrastructure reference for cluster %#q", key.ClusterID(&cl)))

		err = r.k8sClient.CtrlClient().Get(ctx, key.ClusterInfraRef(cl), cr)
		if errors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find latest infrastructure reference for cluster %#q", key.ClusterID(&cl)))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found latest infrastructure reference for cluster %#q", key.ClusterID(&cl)))
	}

	updatedCR := r.computeDeleteClusterStatusConditions(ctx, cr)

	if !reflect.DeepEqual(cr, updatedCR) {
		{
			r.logger.LogCtx(ctx, "level", "debug", "message", "updating cluster status")

			err := r.k8sClient.CtrlClient().Status().Update(ctx, updatedCR)
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

func (r *Resource) computeDeleteClusterStatusConditions(ctx context.Context, obj infrastructurev1alpha2.CommonClusterObject) infrastructurev1alpha2.CommonClusterObject {
	cr := (obj.DeepCopyObject()).(infrastructurev1alpha2.CommonClusterObject)

	status := cr.GetCommonClusterStatus()

	// On Deletion we always add the deleting status condition.
	// We skip adding the condition if it's already set.
	{
		notDeleting := !status.HasDeletingCondition()
		if notDeleting {
			status.Conditions = status.WithDeletingCondition()
			cr.SetCommonClusterStatus(status)
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", infrastructurev1alpha2.ClusterStatusConditionDeleting))
		}
	}

	return cr
}