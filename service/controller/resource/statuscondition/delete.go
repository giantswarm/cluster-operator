package statuscondition

import (
	"context"
	"fmt"
	"reflect"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/v4/pkg/controller/context/reconciliationcanceledcontext"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr := r.newCommonClusterObjectFunc()
	{
		cl, err := key.ToCluster(obj)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "finding latest infrastructure reference for cluster %#q", key.ClusterID(&cl))

		err = r.k8sClient.CtrlClient().Get(ctx, key.ObjRefToNamespacedName(key.ObjRefFromCluster(cl)), cr)
		if errors.IsNotFound(err) {
			r.logger.Debugf(ctx, "did not find latest infrastructure reference for cluster %#q", key.ClusterID(&cl))
			r.logger.Debugf(ctx, "canceling resource")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "found latest infrastructure reference for cluster %#q", key.ClusterID(&cl))
	}

	updatedCR := r.computeDeleteClusterStatusConditions(ctx, cr)

	if !reflect.DeepEqual(cr, updatedCR) {
		{
			r.logger.Debugf(ctx, "updating cluster status")

			err := r.k8sClient.CtrlClient().Status().Update(ctx, updatedCR)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.Debugf(ctx, "updated cluster status")
		}

		{
			r.logger.Debugf(ctx, "canceling reconciliation")
			reconciliationcanceledcontext.SetCanceled(ctx)
			r.logger.Debugf(ctx, "keeping finalizers")
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
