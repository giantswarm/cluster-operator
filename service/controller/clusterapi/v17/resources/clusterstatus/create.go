package clusterstatus

import (
	"context"
	"fmt"
	"reflect"

	"github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v17/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v17/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring cluster status is up to date")

	currentStatus := r.commonClusterStatusAccessor.GetCommonClusterStatus(cr)

	updatedStatus, err := r.ensureClusterHasID(ctx, cr, currentStatus)
	if err != nil {
		return microerror.Mask(err)
	}

	// Ensure that cluster has cluster ID label.
	cr.Labels[label.Cluster] = updatedStatus.ID

	var nodes []corev1.Node
	var tenantAPIAvailable bool
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding nodes of tenant cluster")

		l, err := cc.Client.TenantCluster.K8s.CoreV1().Nodes().List(metav1.ListOptions{})
		if tenant.IsAPINotAvailable(err) {
			// Similarly to above timeout, during cluster creation it is
			// expected that API is not available.
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant API not available")
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			tenantAPIAvailable = true
			nodes = l.Items

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d nodes from tenant cluster", len(nodes)))
		}
	}

	var machineDeployments []cmav1alpha1.MachineDeployment
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding MachineDeployments for tenant cluster")

		machineDeployments, err = r.getMachineDeployments(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d MachineDeployments for tenant cluster", len(machineDeployments)))
	}

	updatedStatus = r.computeClusterConditions(ctx, cr, updatedStatus, tenantAPIAvailable, nodes, machineDeployments)

	if !reflect.DeepEqual(currentStatus, updatedStatus) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating cluster status")

		cr = r.commonClusterStatusAccessor.SetCommonClusterStatus(cr, updatedStatus)
		_, err = r.cmaClient.ClusterV1alpha1().Clusters(cr.Namespace).Update(&cr)
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

// TODO have separate resource for this.
//
//     always execute
//
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

	return status, microerror.Maskf(notFoundError, "cluster ID")
}

// TODO have separate resource for this.
//
//     cancel when client is not in controller context
//
func (r *Resource) computeClusterConditions(ctx context.Context, cluster cmav1alpha1.Cluster, clusterStatus v1alpha1.CommonClusterStatus, tenantAPIAvailable bool, nodes []corev1.Node, machineDeployments []cmav1alpha1.MachineDeployment) v1alpha1.CommonClusterStatus {
	currentVersion := clusterStatus.LatestVersion()
	desiredVersion := key.ClusterReleaseVersion(cluster)

	// Count total number of all workers and number of Ready workers that
	// belong to this cluster.
	var desiredWorkers int
	var readyWorkers int
	{
		for _, md := range machineDeployments {
			desiredWorkers += int(md.Status.Replicas)
		}

		for _, n := range nodes {
			for _, c := range n.Status.Conditions {
				if c.Type == corev1.NodeReady && c.Status == corev1.ConditionTrue {
					readyWorkers++
				}
			}
		}
	}

	// After initialization the most likely implication is the tenant cluster
	// being in a creation status. In case no other conditions are given and no
	// versions are set, we set the tenant cluster status to a creating
	// condition.
	{
		notCreating := !clusterStatus.HasCreatingCondition()
		conditionsEmpty := len(clusterStatus.Conditions) == 0
		versionsEmpty := len(clusterStatus.Versions) == 0

		if notCreating && conditionsEmpty && versionsEmpty {
			clusterStatus.Conditions = clusterStatus.WithCreatingCondition()
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", v1alpha1.ClusterStatusConditionCreating))
		}
	}

	// Once the tenant cluster is created we set the according status condition so
	// the cluster status reflects the transitioning from creating to created.
	{
		isCreating := clusterStatus.HasCreatingCondition()
		notCreated := !clusterStatus.HasCreatedCondition()
		sameCount := readyWorkers == desiredWorkers
		sameVersion := allNodesHaveVersion(nodes, desiredVersion)

		if isCreating && notCreated && sameCount && sameVersion && tenantAPIAvailable {
			clusterStatus.Conditions = clusterStatus.WithCreatedCondition()
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", v1alpha1.ClusterStatusConditionCreated))
		}
	}

	// When we notice the current and the desired tenant cluster version differs,
	// an update is about to be processed. So we set the status condition
	// indicating the tenant cluster is updating now.
	{
		isCreated := clusterStatus.HasCreatedCondition()
		notUpdating := !clusterStatus.HasUpdatingCondition()
		versionDiffers := currentVersion != "" && currentVersion != desiredVersion

		if isCreated && notUpdating && versionDiffers {
			clusterStatus.Conditions = clusterStatus.WithUpdatingCondition()
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", v1alpha1.ClusterStatusConditionUpdating))
		}
	}

	// Set the status cluster condition to updated when an update successfully
	// took place. Precondition for this is the tenant cluster is updating and all
	// nodes being known and all nodes having the same versions.
	{
		isUpdating := clusterStatus.HasUpdatingCondition()
		notUpdated := !clusterStatus.HasUpdatedCondition()
		sameCount := readyWorkers != 0 && readyWorkers == desiredWorkers
		sameVersion := allNodesHaveVersion(nodes, desiredVersion)

		if isUpdating && notUpdated && sameCount && sameVersion && tenantAPIAvailable {
			clusterStatus.Conditions = clusterStatus.WithUpdatedCondition()
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", v1alpha1.ClusterStatusConditionUpdated))
		}
	}

	// Check all node versions held by the cluster status and add the version the
	// tenant cluster successfully migrated to, to the historical list of versions.
	{
		hasTransitioned := clusterStatus.HasCreatedCondition() || clusterStatus.HasUpdatedCondition()
		notSet := !clusterStatus.HasVersion(desiredVersion)
		sameCount := readyWorkers != 0 && readyWorkers == desiredWorkers
		sameVersion := allNodesHaveVersion(nodes, desiredVersion)

		if hasTransitioned && notSet && sameCount && sameVersion {
			clusterStatus.Versions = clusterStatus.WithNewVersion(desiredVersion)
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting status versions with new version: %q", desiredVersion))
		}
	}

	return clusterStatus
}

func (r *Resource) getMachineDeployments(ctx context.Context, cluster cmav1alpha1.Cluster) ([]cmav1alpha1.MachineDeployment, error) {
	labelSelector := metav1.AddLabelToSelector(&v1.LabelSelector{}, label.Cluster, key.ClusterID(&cluster))
	listOpts := metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	}

	mdList, err := r.cmaClient.ClusterV1alpha1().MachineDeployments(cluster.Namespace).List(listOpts)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return mdList.Items, nil
}

func allNodesHaveVersion(nodes []corev1.Node, version string) bool {
	if len(nodes) == 0 {
		return false
	}

	for _, n := range nodes {
		v := key.ReleaseVersion(&n)
		if v != version {
			return false
		}
	}

	return true
}
