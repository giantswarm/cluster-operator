package clusterstatus

import (
	"context"
	"fmt"
	"reflect"

	"github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v17/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cluster, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring cluster status is up to date")

	currentStatus := r.commonClusterStatusAccessor.GetCommonClusterStatus(cluster)

	updatedStatus, err := r.ensureClusterHasID(ctx, cluster, currentStatus)
	if err != nil {
		return microerror.Mask(err)
	}

	// Ensure that cluster has cluster ID label.
	cluster.Labels[label.Cluster] = updatedStatus.ID

	updatedStatus = r.computeClusterConditions(ctx, cluster, updatedStatus)

	updatedStatus = r.computeClusterVersion(ctx, cluster, updatedStatus)

	if !reflect.DeepEqual(currentStatus, updatedStatus) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating cluster status")

		cluster = r.commonClusterStatusAccessor.SetCommonClusterStatus(cluster, updatedStatus)
		_, err = r.cmaClient.ClusterV1alpha1().Clusters(cluster.Namespace).Update(&cluster)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated cluster status")

		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)

		return nil
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "ensured cluster status is up to date")

	return nil
}

func (r *Resource) ensureClusterHasID(ctx context.Context, cluster cmav1alpha1.Cluster, status v1alpha1.CommonClusterStatus) (v1alpha1.CommonClusterStatus, error) {
	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if cluster status has ID")

	if status.ID != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found out cluster has ID %s", status.ID))
		return status, nil
	}

	clusterID := cluster.Labels[label.Cluster]
	if clusterID != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found out cluster has cluster label containing %s; reusing that", status.ID))
		status.ID = clusterID
		return status, nil
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "did not found cluster ID; generating one")

	panic("tuommaki: Implement cluster ID generation.")

	//return status, nil
}

func (r *Resource) computeClusterConditions(ctx context.Context, cluster cmav1alpha1.Cluster, status v1alpha1.CommonClusterStatus) v1alpha1.CommonClusterStatus {
	return status
}

func (r *Resource) computeClusterVersion(ctx context.Context, cluster cmav1alpha1.Cluster, status v1alpha1.CommonClusterStatus) v1alpha1.CommonClusterStatus {
	return status
}
