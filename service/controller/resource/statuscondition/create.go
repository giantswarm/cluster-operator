package statuscondition

import (
	"context"
	"fmt"
	"reflect"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/reconciliationcanceledcontext"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	apiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/v3/pkg/label"
	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
	"github.com/giantswarm/cluster-operator/v3/service/internal/tenantclient"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr := r.newCommonClusterObjectFunc()
	var uc infrastructurev1alpha3.CommonClusterObject
	{
		r.logger.Debugf(ctx, "finding latest cluster")

		cl, err := key.ToCluster(obj)
		if err != nil {
			return microerror.Mask(err)
		}

		err = r.k8sClient.CtrlClient().Get(ctx, key.ObjRefToNamespacedName(key.ObjRefFromCluster(cl)), cr)
		if err != nil {
			return microerror.Mask(err)
		}

		uc = cr.DeepCopyObject().(infrastructurev1alpha3.CommonClusterObject)

		r.logger.Debugf(ctx, "found latest cluster")
	}

	var cl apiv1beta1.Cluster
	{
		r.logger.Debugf(ctx, "finding cluster")

		c, err := key.ToCluster(obj)
		if err != nil {
			return microerror.Mask(err)
		}

		err = r.k8sClient.CtrlClient().Get(ctx, types.NamespacedName{Name: c.GetName(), Namespace: c.GetNamespace()}, &cl)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "found cluster")
	}

	tenantClient, err := r.tenantClient.K8sClient(ctx, cr)
	if tenantclient.IsNotAvailable(err) {
		r.logger.Debugf(ctx, "tenant client is not available yet")
	} else if err != nil {
		return microerror.Mask(err)
	}

	var nodes []corev1.Node
	if tenantClient != nil {
		r.logger.Debugf(ctx, "finding nodes of tenant cluster")
		l, err := tenantClient.K8sClient().CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if tenant.IsAPINotAvailable(err) {
			// During cluster creation / upgrade the tenant API is naturally not
			// available but this resource must still continue execution as that's
			// when `Creating` and `Upgrading` conditions may need to be applied.
			r.logger.Debugf(ctx, "tenant API not available yet")
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			nodes = l.Items

			r.logger.Debugf(ctx, "found %d nodes from tenant cluster", len(nodes))
		}
	}

	cpList := &infrastructurev1alpha3.G8sControlPlaneList{}
	{
		r.logger.Debugf(ctx, "finding G8sControlplane for tenant cluster")

		err = r.k8sClient.CtrlClient().List(
			ctx,
			cpList,
			client.InNamespace(cr.GetNamespace()),
			client.MatchingLabels{label.Cluster: key.ClusterID(cr)},
		)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "found %d G8sControlplane for tenant cluster", len(cpList.Items))
	}

	mdList := &apiv1beta1.MachineDeploymentList{}
	{
		r.logger.Debugf(ctx, "finding MachineDeployments for tenant cluster")

		err = r.k8sClient.CtrlClient().List(
			ctx,
			mdList,
			client.InNamespace(cr.GetNamespace()),
			client.MatchingLabels{label.Cluster: key.ClusterID(cr)},
		)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "found %d MachineDeployments for tenant cluster", len(mdList.Items))
	}

	err = r.computeClusterStatusConditions(ctx, cl, uc, nodes, cpList.Items, mdList.Items)
	if err != nil {
		return microerror.Mask(err)
	}

	if !reflect.DeepEqual(cr.GetCommonClusterStatus(), uc.GetCommonClusterStatus()) {
		r.logger.Debugf(ctx, "updating cluster status")

		err := r.k8sClient.CtrlClient().Status().Update(ctx, uc)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "updated cluster status")

		r.logger.Debugf(ctx, "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)

		return nil
	}

	return nil
}

func (r *Resource) computeClusterStatusConditions(ctx context.Context, cl apiv1beta1.Cluster, cr infrastructurev1alpha3.CommonClusterObject, nodes []corev1.Node, controlPlanes []infrastructurev1alpha3.G8sControlPlane, machineDeployments []apiv1beta1.MachineDeployment) error {
	var desiredVersion string
	var nodesReady bool

	desiredVersion, err := r.getDesiredVersion(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}
	{
		providerOperatorVersionLabel := fmt.Sprintf("%s-operator.giantswarm.io/version", r.provider)
		sameVersion := allNodesHaveVersion(nodes, desiredVersion, providerOperatorVersionLabel)
		sameMasterCount := allMasterNodesReady(controlPlanes)
		sameWorkerCount := allWorkerNodesReady(machineDeployments)

		nodesReady = sameMasterCount && sameWorkerCount && sameVersion
	}

	return r.writeClusterStatusConditions(ctx, cl, cr, nodesReady, desiredVersion)
}

func (r *Resource) writeClusterStatusConditions(ctx context.Context, cl apiv1beta1.Cluster, cr infrastructurev1alpha3.CommonClusterObject, nodesReady bool, desiredVersion string) error {
	status := cr.GetCommonClusterStatus()

	// After initialization the most likely implication is the tenant cluster
	// being in a creation status. In case no other conditions are given and no
	// versions are set, we set the tenant cluster status to a creating
	// condition.
	if computeCreatingCondition(status) {
		status.Conditions = status.WithCreatingCondition()
		r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", infrastructurev1alpha3.ClusterStatusConditionCreating))
		r.event.Emit(ctx, &cl, "ClusterInCreation", fmt.Sprintf("cluster creation is in condition %s", infrastructurev1alpha3.ClusterStatusConditionCreating))
	}

	// Once the tenant cluster is created we set the according status condition so
	// the cluster status reflects the transitioning from creating to created.
	if computeCreatedCondition(status, nodesReady) {
		status.Conditions = status.WithCreatedCondition()
		r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", infrastructurev1alpha3.ClusterStatusConditionCreated))
		r.event.Emit(ctx, &cl, "ClusterCreated", fmt.Sprintf("cluster is in condition %s", infrastructurev1alpha3.ClusterStatusConditionCreated))
	}

	// When we notice the current and the desired tenant cluster version differs,
	// an update is about to be processed. So we set the status condition
	// indicating the tenant cluster is updating now.
	if computeUpdatingCondition(status, desiredVersion) {
		status.Conditions = status.WithUpdatingCondition()
		r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", infrastructurev1alpha3.ClusterStatusConditionUpdating))
		r.event.Emit(ctx, &cl, "ClusterIsUpdating", fmt.Sprintf("cluster is in condition %s", infrastructurev1alpha3.ClusterStatusConditionUpdating))
	}

	// Set the status cluster condition to updated when an update successfully
	// took place. Precondition for this is the tenant cluster is updating and all
	// nodes being known and all nodes having the same versions.
	if computeUpdatedCondition(status, nodesReady) {
		status.Conditions = status.WithUpdatedCondition()
		r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", infrastructurev1alpha3.ClusterStatusConditionUpdated))
		r.event.Emit(ctx, &cl, "ClusterUpdated", fmt.Sprintf("cluster is in condition %s", infrastructurev1alpha3.ClusterStatusConditionUpdated))
	}

	// Check all node versions held by the cluster status and add the version the
	// tenant cluster successfully migrated to, to the historical list of versions.
	if computeVersionChange(status, nodesReady, desiredVersion) {
		status.Versions = status.WithNewVersion(desiredVersion)
		r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting status versions with new version: %q", desiredVersion))
		r.event.Emit(ctx, &cl, "ClusterVersionUpdated", fmt.Sprintf("cluster status set with new version: %q", desiredVersion))
	}
	cr.SetCommonClusterStatus(status)
	return nil
}

func (r *Resource) getDesiredVersion(ctx context.Context, cr infrastructurev1alpha3.CommonClusterObject) (string, error) {
	componentVersions, err := r.releaseVersion.ComponentVersion(ctx, cr)
	if err != nil {
		return "", microerror.Mask(err)
	}

	providerOperator := fmt.Sprintf("%s-operator", r.provider)

	providerComponent := componentVersions[providerOperator]
	desiredVersion := providerComponent.Version
	if desiredVersion == "" {
		return "", microerror.Maskf(notFoundError, "component version not found for %#q", providerOperator)
	}

	return desiredVersion, nil
}

func allMasterNodesReady(controlPlanes []infrastructurev1alpha3.G8sControlPlane) bool {
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
	return readyMasterReplicas == desiredMasterReplicas
}

func allWorkerNodesReady(machineDeployments []apiv1beta1.MachineDeployment) bool {
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
	return readyWorkerReplicas == desiredWorkerReplicas
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

func computeCreatingCondition(status infrastructurev1alpha3.CommonClusterStatus) bool {
	notCreating := !status.HasCreatingCondition()
	conditionsEmpty := len(status.Conditions) == 0
	versionsEmpty := len(status.Versions) == 0

	return notCreating && conditionsEmpty && versionsEmpty
}

func computeCreatedCondition(status infrastructurev1alpha3.CommonClusterStatus, nodesReady bool) bool {
	isCreating := status.HasCreatingCondition()
	notCreated := !status.HasCreatedCondition()

	return isCreating && notCreated && nodesReady
}

func computeUpdatingCondition(status infrastructurev1alpha3.CommonClusterStatus, desiredVersion string) bool {
	currentVersion := status.LatestVersion()
	isCreated := status.HasCreatedCondition()
	notUpdating := status.LatestCondition() != infrastructurev1alpha3.ClusterStatusConditionUpdating
	versionDiffers := currentVersion != "" && currentVersion != desiredVersion

	return isCreated && notUpdating && versionDiffers
}

func computeUpdatedCondition(status infrastructurev1alpha3.CommonClusterStatus, nodesReady bool) bool {
	isUpdating := status.LatestCondition() == infrastructurev1alpha3.ClusterStatusConditionUpdating
	notUpdated := status.LatestCondition() != infrastructurev1alpha3.ClusterStatusConditionUpdated

	return isUpdating && notUpdated && nodesReady
}

func computeVersionChange(status infrastructurev1alpha3.CommonClusterStatus, nodesReady bool, desiredVersion string) bool {
	isCreated := status.LatestCondition() == infrastructurev1alpha3.ClusterStatusConditionCreated
	isUpdated := status.LatestCondition() == infrastructurev1alpha3.ClusterStatusConditionUpdated
	hasTransitioned := isCreated || isUpdated
	notSet := status.LatestVersion() != desiredVersion

	return hasTransitioned && notSet && nodesReady
}
