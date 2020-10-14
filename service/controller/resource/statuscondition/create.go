package statuscondition

import (
	"context"
	"fmt"
	"reflect"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v2/pkg/controller/context/reconciliationcanceledcontext"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/v3/pkg/label"
	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr := r.newCommonClusterObjectFunc()
	var uc infrastructurev1alpha2.CommonClusterObject
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding latest cluster")

		cl, err := key.ToCluster(obj)
		if err != nil {
			return microerror.Mask(err)
		}

		err = r.k8sClient.CtrlClient().Get(ctx, key.ObjRefToNamespacedName(key.ObjRefFromCluster(cl)), cr)
		if err != nil {
			return microerror.Mask(err)
		}

		uc = cr.DeepCopyObject().(infrastructurev1alpha2.CommonClusterObject)

		r.logger.LogCtx(ctx, "level", "debug", "message", "found latest cluster")
	}

	var cl apiv1alpha2.Cluster
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding cluster")

		c, err := key.ToCluster(obj)
		if err != nil {
			return microerror.Mask(err)
		}

		err = r.k8sClient.CtrlClient().Get(ctx, types.NamespacedName{Name: c.GetName(), Namespace: c.GetNamespace()}, &cl)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found cluster")
	}

	tenantClient, err := r.tenantClient.K8sClient(ctx, cr)
	if err != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant client not available yet", "stack", microerror.JSON(err))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	var nodes []corev1.Node
	if tenantClient != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding nodes of tenant cluster")
		l, err := tenantClient.K8sClient().CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			// During cluster creation / upgrade the tenant API is naturally not
			// available but this resource must still continue execution as that's
			// when `Creating` and `Upgrading` conditions may need to be applied.
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant API not available yet", "stack", microerror.JSON(err))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		} else {
			nodes = l.Items

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d nodes from tenant cluster", len(nodes)))
		}
	}

	cpList := &infrastructurev1alpha2.G8sControlPlaneList{}
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding G8sControlplane for tenant cluster")

		err = r.k8sClient.CtrlClient().List(
			ctx,
			cpList,
			client.InNamespace(cr.GetNamespace()),
			client.MatchingLabels{label.Cluster: key.ClusterID(cr)},
		)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d G8sControlplane for tenant cluster", len(cpList.Items)))
	}

	mdList := &apiv1alpha2.MachineDeploymentList{}
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding MachineDeployments for tenant cluster")

		err = r.k8sClient.CtrlClient().List(
			ctx,
			mdList,
			client.InNamespace(cr.GetNamespace()),
			client.MatchingLabels{label.Cluster: key.ClusterID(cr)},
		)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d MachineDeployments for tenant cluster", len(mdList.Items)))
	}

	err = r.computeCreateClusterStatusConditions(ctx, cl, uc, nodes, cpList.Items, mdList.Items)
	if err != nil {
		return microerror.Mask(err)
	}

	if !reflect.DeepEqual(cr.GetCommonClusterStatus(), uc.GetCommonClusterStatus()) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating cluster status")

		err := r.k8sClient.CtrlClient().Status().Update(ctx, uc)
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

func (r *Resource) computeCreateClusterStatusConditions(ctx context.Context, cl apiv1alpha2.Cluster, cr infrastructurev1alpha2.CommonClusterObject, nodes []corev1.Node, controlPlanes []infrastructurev1alpha2.G8sControlPlane, machineDeployments []apiv1alpha2.MachineDeployment) error {
	componentVersions, err := r.releaseVersion.ComponentVersion(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	providerOperator := fmt.Sprintf("%s-operator", r.provider)
	providerOperatorVersionLabel := fmt.Sprintf("%s-operator.giantswarm.io/version", r.provider)

	status := cr.GetCommonClusterStatus()

	var currentVersion string
	var desiredVersion string
	{
		currentVersion = status.LatestVersion()
		desiredVersion = componentVersions[providerOperator]
	}

	// Count total number of all masters and number of ready masters that
	// belong to this cluster.
	var desiredMasterReplicas int
	var readyMasterReplicas int
	{
		for _, cp := range controlPlanes {
			desiredMasterReplicas += int(cp.Status.Replicas)
		}

		for _, cp := range controlPlanes {
			readyMasterReplicas += int(cp.Status.ReadyReplicas)
		}
	}

	// Count total number of all workers and number of Ready workers that
	// belong to this cluster.
	var desiredWorkerReplicas int
	var readyWorkerReplicas int
	{
		for _, md := range machineDeployments {
			desiredWorkerReplicas += int(md.Status.Replicas)
		}

		for _, md := range machineDeployments {
			readyWorkerReplicas += int(md.Status.ReadyReplicas)
		}
	}

	// After initialization the most likely implication is the tenant cluster
	// being in a creation status. In case no other conditions are given and no
	// versions are set, we set the tenant cluster status to a creating
	// condition.
	{
		notCreating := !status.HasCreatingCondition()
		conditionsEmpty := len(status.Conditions) == 0
		versionsEmpty := len(status.Versions) == 0

		if notCreating && conditionsEmpty && versionsEmpty {
			status.Conditions = status.WithCreatingCondition()
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", infrastructurev1alpha2.ClusterStatusConditionCreating))
			r.event.Emit(ctx, &cl, "ClusterInCreation", fmt.Sprintf("cluster creation is in condition %s", infrastructurev1alpha2.ClusterStatusConditionCreating))
		}
	}

	// Once the tenant cluster is created we set the according status condition so
	// the cluster status reflects the transitioning from creating to created.
	{
		isCreating := status.HasCreatingCondition()
		notCreated := !status.HasCreatedCondition()
		sameMasterCount := readyMasterReplicas == desiredMasterReplicas
		sameWorkerCount := readyWorkerReplicas == desiredWorkerReplicas
		sameVersion := allNodesHaveVersion(nodes, desiredVersion, providerOperatorVersionLabel)

		if isCreating && notCreated && sameMasterCount && sameWorkerCount && sameVersion {
			status.Conditions = status.WithCreatedCondition()
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", infrastructurev1alpha2.ClusterStatusConditionCreated))
			r.event.Emit(ctx, &cl, "ClusterCreated", fmt.Sprintf("cluster is in condition %s", infrastructurev1alpha2.ClusterStatusConditionCreated))
		}
	}

	// When we notice the current and the desired tenant cluster version differs,
	// an update is about to be processed. So we set the status condition
	// indicating the tenant cluster is updating now.
	{
		isCreated := status.HasCreatedCondition()
		notUpdating := !status.HasUpdatingCondition()
		versionDiffers := currentVersion != "" && currentVersion != desiredVersion

		if isCreated && notUpdating && versionDiffers {
			status.Conditions = status.WithUpdatingCondition()
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", infrastructurev1alpha2.ClusterStatusConditionUpdating))
			r.event.Emit(ctx, &cl, "ClusterIsUpdating", fmt.Sprintf("cluster is in condition %s", infrastructurev1alpha2.ClusterStatusConditionUpdating))
		}
	}

	// Set the status cluster condition to updated when an update successfully
	// took place. Precondition for this is the tenant cluster is updating and all
	// nodes being known and all nodes having the same versions.
	{
		isUpdating := status.HasUpdatingCondition()
		notUpdated := !status.HasUpdatedCondition()
		sameMasterCount := readyMasterReplicas != 0 && readyMasterReplicas == desiredMasterReplicas
		sameWorkerCount := readyWorkerReplicas != 0 && readyWorkerReplicas == desiredWorkerReplicas
		sameVersion := allNodesHaveVersion(nodes, desiredVersion, providerOperatorVersionLabel)

		if isUpdating && notUpdated && sameMasterCount && sameWorkerCount && sameVersion {
			status.Conditions = status.WithUpdatedCondition()
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", infrastructurev1alpha2.ClusterStatusConditionUpdated))
			r.event.Emit(ctx, &cl, "ClusterUpdated", fmt.Sprintf("cluster is in condition %s", infrastructurev1alpha2.ClusterStatusConditionUpdated))
		}
	}

	// Check all node versions held by the cluster status and add the version the
	// tenant cluster successfully migrated to, to the historical list of versions.
	{
		hasTransitioned := status.HasCreatedCondition() || status.HasUpdatedCondition()
		notSet := !status.HasVersion(desiredVersion)
		sameMasterCount := readyMasterReplicas != 0 && readyMasterReplicas == desiredMasterReplicas
		sameWorkerCount := readyWorkerReplicas != 0 && readyWorkerReplicas == desiredWorkerReplicas
		sameVersion := allNodesHaveVersion(nodes, desiredVersion, providerOperatorVersionLabel)

		if hasTransitioned && notSet && sameMasterCount && sameWorkerCount && sameVersion {
			status.Versions = status.WithNewVersion(desiredVersion)
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting status versions with new version: %q", desiredVersion))
			r.event.Emit(ctx, &cl, "ClusterVersionUpdated", fmt.Sprintf("cluster status set with new version: %q", desiredVersion))
		}
	}

	cr.SetCommonClusterStatus(status)

	return nil
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
