package clusterstatus

import (
	"context"
	"fmt"
	"reflect"

	g8sv1alpha "github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v21/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v21/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	old, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	if cc.Client.TenantCluster.K8s == nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster clients not available in controller context")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	}

	var cr cmav1alpha1.Cluster
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding latest cluster")

		cl, err := r.cmaClient.ClusterV1alpha1().Clusters(old.Namespace).Get(old.Name, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		cr = *cl

		r.logger.LogCtx(ctx, "level", "debug", "message", "found latest cluster")
	}

	var nodes []corev1.Node
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding nodes of tenant cluster")

		l, err := cc.Client.TenantCluster.K8s.CoreV1().Nodes().List(metav1.ListOptions{})
		if tenant.IsAPINotAvailable(err) {
			// During cluster creation / upgrade the tenant API is naturally not
			// available but this resource must still continue execution as that's
			// when `Creating` and `Upgrading` conditions may need to be applied.
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant API not available")
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			nodes = l.Items

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d nodes from tenant cluster", len(nodes)))
		}
	}

	var machineDeployments []cmav1alpha1.MachineDeployment
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding MachineDeployments for tenant cluster")

		l := metav1.AddLabelToSelector(
			&v1.LabelSelector{},
			label.Cluster,
			key.ClusterID(&cr),
		)
		o := metav1.ListOptions{
			LabelSelector: labels.Set(l.MatchLabels).String(),
		}

		list, err := r.cmaClient.ClusterV1alpha1().MachineDeployments(cr.Namespace).List(o)
		if err != nil {
			return microerror.Mask(err)
		}

		machineDeployments = list.Items

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d MachineDeployments for tenant cluster", len(machineDeployments)))
	}

	updatedStatus := r.computeClusterConditions(ctx, cr, r.accessor.GetCommonClusterStatus(cr), nodes, machineDeployments)

	if !reflect.DeepEqual(r.accessor.GetCommonClusterStatus(cr), updatedStatus) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating cluster status")

		cr = r.accessor.SetCommonClusterStatus(cr, updatedStatus)

		_, err := r.cmaClient.ClusterV1alpha1().Clusters(cr.Namespace).UpdateStatus(&cr)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated cluster status")

		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)

		return nil
	}

	return nil
}

func (r *Resource) computeClusterConditions(ctx context.Context, cluster cmav1alpha1.Cluster, clusterStatus g8sv1alpha.CommonClusterStatus, nodes []corev1.Node, machineDeployments []cmav1alpha1.MachineDeployment) g8sv1alpha.CommonClusterStatus {
	var currentVersion string
	var desiredVersion string
	{
		currentVersion = clusterStatus.LatestVersion()
		desiredVersion = key.ReleaseVersion(&cluster)
	}

	// Count total number of all workers and number of Ready workers that
	// belong to this cluster.
	var desiredReplicas int
	var readyReplicas int
	{
		for _, md := range machineDeployments {
			desiredReplicas += int(md.Status.Replicas)
		}

		for _, md := range machineDeployments {
			readyReplicas += int(md.Status.ReadyReplicas)
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
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", g8sv1alpha.ClusterStatusConditionCreating))
		}
	}

	// Once the tenant cluster is created we set the according status condition so
	// the cluster status reflects the transitioning from creating to created.
	{
		isCreating := clusterStatus.HasCreatingCondition()
		notCreated := !clusterStatus.HasCreatedCondition()
		sameCount := readyReplicas == desiredReplicas
		sameVersion := allNodesHaveVersion(nodes, desiredVersion)

		if isCreating && notCreated && sameCount && sameVersion {
			clusterStatus.Conditions = clusterStatus.WithCreatedCondition()
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", g8sv1alpha.ClusterStatusConditionCreated))
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
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", g8sv1alpha.ClusterStatusConditionUpdating))
		}
	}

	// Set the status cluster condition to updated when an update successfully
	// took place. Precondition for this is the tenant cluster is updating and all
	// nodes being known and all nodes having the same versions.
	{
		isUpdating := clusterStatus.HasUpdatingCondition()
		notUpdated := !clusterStatus.HasUpdatedCondition()
		sameCount := readyReplicas != 0 && readyReplicas == desiredReplicas
		sameVersion := allNodesHaveVersion(nodes, desiredVersion)

		if isUpdating && notUpdated && sameCount && sameVersion {
			clusterStatus.Conditions = clusterStatus.WithUpdatedCondition()
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", g8sv1alpha.ClusterStatusConditionUpdated))
		}
	}

	// Check all node versions held by the cluster status and add the version the
	// tenant cluster successfully migrated to, to the historical list of versions.
	{
		hasTransitioned := clusterStatus.HasCreatedCondition() || clusterStatus.HasUpdatedCondition()
		notSet := !clusterStatus.HasVersion(desiredVersion)
		sameCount := readyReplicas != 0 && readyReplicas == desiredReplicas
		sameVersion := allNodesHaveVersion(nodes, desiredVersion)

		if hasTransitioned && notSet && sameCount && sameVersion {
			clusterStatus.Versions = clusterStatus.WithNewVersion(desiredVersion)
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting status versions with new version: %q", desiredVersion))
		}
	}

	return clusterStatus
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
