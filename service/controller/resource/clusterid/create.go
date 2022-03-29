package clusterid

import (
	"context"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/reconciliationcanceledcontext"

	"github.com/giantswarm/cluster-operator/v4/pkg/label"
	"github.com/giantswarm/cluster-operator/v4/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	if r.provider != label.ProviderAWS {
		r.logger.Debugf(ctx, "provider is %q, only supported provider for %q resource is aws", r.provider, r.Name())
		r.logger.Debugf(ctx, "canceling resource")
		return nil
	}

	cr := r.newCommonClusterObjectFunc()
	var status infrastructurev1alpha3.CommonClusterStatus
	{
		cl, err := key.ToCluster(obj)
		if err != nil {
			return microerror.Mask(err)
		}

		err = r.k8sClient.CtrlClient().Get(ctx, key.ObjRefToNamespacedName(key.ObjRefFromCluster(cl)), cr)
		if err != nil {
			return microerror.Mask(err)
		}

		status = cr.GetCommonClusterStatus()
	}

	{
		if status.ID != "" {
			r.logger.Debugf(ctx, "cluster %#q has cluster id in status", cr.GetName())
			r.logger.Debugf(ctx, "canceling resource")

			return nil
		}

		if key.ClusterID(cr) == "" {
			r.logger.Debugf(ctx, "cluster %#q misses cluster id in labels", cr.GetName())
			r.logger.Debugf(ctx, "canceling resource")
			return nil
		}
	}

	{
		r.logger.Debugf(ctx, "updating cluster status")

		status.ID = key.ClusterID(cr)

		cr.SetCommonClusterStatus(status)

		err := r.k8sClient.CtrlClient().Status().Update(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "updated cluster status")

		// All further resources require cluster ID to be present in the status so
		// it makes sense to cancel whole CR reconciliation here and start from the
		// beginning.
		r.logger.Debugf(ctx, "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
	}

	return nil
}
