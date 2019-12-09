package clusterid

import (
	"context"
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/cluster-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	statusReader := &infrastructurev1alpha2.StatusReader{}
	err = r.k8sClient.CtrlClient().Get(ctx, types.NamespacedName{Name: cr.Spec.InfrastructureRef.Name, Namespace: cr.Spec.InfrastructureRef.Namespace}, statusReader)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		if statusReader.Status.Cluster.ID != "" {
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

		statusReader.Status.Cluster.ID = key.ClusterID(&cr)

		err = r.k8sClient.CtrlClient().Status().Update(ctx, statusReader)
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
