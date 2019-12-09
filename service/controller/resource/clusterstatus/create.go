package clusterstatus

import (
	"context"
	"fmt"
	"reflect"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	if cc.Client.TenantCluster.K8s == nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster clients not available in controller context")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	}

	cr := &apiv1alpha2.Cluster{}
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding latest cluster")

		cl, err := key.ToCluster(obj)
		if err != nil {
			return microerror.Mask(err)
		}

		err = r.k8sClient.CtrlClient().Get(ctx, key.InfrastructureRef(cl), cr)
		if err != nil {
			return microerror.Mask(err)
		}

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

	machineDeployments := &apiv1alpha2.MachineDeploymentList{}
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding MachineDeployments for tenant cluster")

		err = r.k8sClient.CtrlClient().List(
			ctx,
			machineDeployments,
			client.InNamespace(cr.Namespace),
			client.MatchingLabels{label.Cluster: key.ClusterID(cr)},
		)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d MachineDeployments for tenant cluster", len(machineDeployments.Items)))
	}

	updatedStatus := r.computeClusterConditions(ctx, cc, cr, r.accessor.GetCommonClusterStatus(cr), nodes, machineDeployments)

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

func (r *Resource) computeClusterConditions(ctx context.Context, cc *controllercontext.Context, cluster apiv1alpha2.Cluster, clusterStatus infrastructurev1alpha2.CommonClusterStatus, nodes []corev1.Node, machineDeployments []apiv1alpha2.MachineDeployment) infrastructurev1alpha2.CommonClusterStatus {
	providerOperatorVersionLabel := fmt.Sprintf("%s-operator.giantswarm.io/version", r.provider)

	var currentVersion string
	var desiredVersion string
	{
		currentVersion = clusterStatus.LatestVersion()
		desiredVersion = cc.Status.Versions[providerOperatorVersionLabel]
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
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", infrastructurev1alpha2.ClusterStatusConditionCreating))
		}
	}

	// Once the tenant cluster is created we set the according status condition so
	// the cluster status reflects the transitioning from creating to created.
	{
		isCreating := clusterStatus.HasCreatingCondition()
		notCreated := !clusterStatus.HasCreatedCondition()
		sameCount := readyReplicas == desiredReplicas
		sameVersion := allNodesHaveVersion(nodes, desiredVersion, providerOperatorVersionLabel)

		if isCreating && notCreated && sameCount && sameVersion {
			clusterStatus.Conditions = clusterStatus.WithCreatedCondition()
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", infrastructurev1alpha2.ClusterStatusConditionCreated))
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
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", infrastructurev1alpha2.ClusterStatusConditionUpdating))
		}
	}

	// Set the status cluster condition to updated when an update successfully
	// took place. Precondition for this is the tenant cluster is updating and all
	// nodes being known and all nodes having the same versions.
	{
		isUpdating := clusterStatus.HasUpdatingCondition()
		notUpdated := !clusterStatus.HasUpdatedCondition()
		sameCount := readyReplicas != 0 && readyReplicas == desiredReplicas
		sameVersion := allNodesHaveVersion(nodes, desiredVersion, providerOperatorVersionLabel)

		if isUpdating && notUpdated && sameCount && sameVersion {
			clusterStatus.Conditions = clusterStatus.WithUpdatedCondition()
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", infrastructurev1alpha2.ClusterStatusConditionUpdated))
		}
	}

	// Check all node versions held by the cluster status and add the version the
	// tenant cluster successfully migrated to, to the historical list of versions.
	{
		hasTransitioned := clusterStatus.HasCreatedCondition() || clusterStatus.HasUpdatedCondition()
		notSet := !clusterStatus.HasVersion(desiredVersion)
		sameCount := readyReplicas != 0 && readyReplicas == desiredReplicas
		sameVersion := allNodesHaveVersion(nodes, desiredVersion, providerOperatorVersionLabel)

		if hasTransitioned && notSet && sameCount && sameVersion {
			clusterStatus.Versions = clusterStatus.WithNewVersion(desiredVersion)
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting status versions with new version: %q", desiredVersion))
		}
	}

	return clusterStatus
}

func allNodesHaveVersion(nodes []corev1.Node, version string, providerOperatorVersionLabel string) bool {
	if len(nodes) == 0 {
		return false
	}

	for _, n := range nodes {
		v := n.Labels[providerOperatorVersionLabel]
		if v != version {
			return false
		}
	}

	return true
}
