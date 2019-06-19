package clusterid

import (
	"context"
	"reflect"

	"github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v17/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring cluster status has cluster ID")

	currentStatus := r.commonClusterStatusAccessor.GetCommonClusterStatus(cr)

	updatedStatus, err := r.ensureClusterHasID(ctx, cr, currentStatus)
	if err != nil {
		return microerror.Mask(err)
	}

	updatedCR := r.commonClusterStatusAccessor.SetCommonClusterStatus(cr, updatedStatus)

	if !reflect.DeepEqual(currentStatus, updatedStatus) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating cluster status")

		_, err = r.cmaClient.ClusterV1alpha1().Clusters(cr.Namespace).Update(&updatedCR)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated cluster status")

		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)

		return nil
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "ensured cluster status has cluster ID")

	return nil
}

func (r *Resource) ensureClusterHasID(ctx context.Context, cluster cmav1alpha1.Cluster, status v1alpha1.CommonClusterStatus) (v1alpha1.CommonClusterStatus, error) {
	if status.ID != "" {
		return status, nil
	}

	clusterID := cluster.Labels[label.Cluster]
	if clusterID != "" {
		status.ID = clusterID
		return status, nil
	}

	return status, microerror.Maskf(notFoundError, "cluster ID")
}
