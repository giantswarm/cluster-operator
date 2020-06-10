package statuscondition

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
	"github.com/giantswarm/cluster-operator/service/controller/key"
	"github.com/giantswarm/cluster-operator/service/internal/tenantclient"
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

	tenantClient, err := r.tenantClient.K8sClient(ctx, cr)
	if tenantclient.IsNotAvailable(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant client is not available yet")

		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
	} else if err != nil {
		return microerror.Mask(err)
	}

	var nodes []corev1.Node
	r.logger.LogCtx(ctx, "level", "debug", "message", "finding nodes of tenant cluster")

	l, err := tenantClient.K8sClient().CoreV1().Nodes().List(metav1.ListOptions{})
	if tenant.IsAPINotAvailable(err) {
		// During cluster creation / upgrade the tenant API is naturally not
		// available but this resource must still continue execution as that's
		// when `Creating` and `Upgrading` conditions may need to be applied.
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant API not available yet")
	} else if err != nil {
		return microerror.Mask(err)
	} else {
		nodes = l.Items

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d nodes from tenant cluster", len(nodes)))
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

	err = r.computeCreateClusterStatusConditions(ctx, uc, nodes, mdList.Items)
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

func (r *Resource) computeCreateClusterStatusConditions(ctx context.Context, cr infrastructurev1alpha2.CommonClusterObject, nodes []corev1.Node, machineDeployments []apiv1alpha2.MachineDeployment) error {
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
		notCreating := !status.HasCreatingCondition()
		conditionsEmpty := len(status.Conditions) == 0
		versionsEmpty := len(status.Versions) == 0

		if notCreating && conditionsEmpty && versionsEmpty {
			status.Conditions = status.WithCreatingCondition()
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", infrastructurev1alpha2.ClusterStatusConditionCreating))
		}
	}

	// Once the tenant cluster is created we set the according status condition so
	// the cluster status reflects the transitioning from creating to created.
	{
		isCreating := status.HasCreatingCondition()
		notCreated := !status.HasCreatedCondition()
		sameCount := readyReplicas == desiredReplicas
		sameVersion := allNodesHaveVersion(nodes, desiredVersion, providerOperatorVersionLabel)

		if isCreating && notCreated && sameCount && sameVersion {
			status.Conditions = status.WithCreatedCondition()
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", infrastructurev1alpha2.ClusterStatusConditionCreated))
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
		}
	}

	// Set the status cluster condition to updated when an update successfully
	// took place. Precondition for this is the tenant cluster is updating and all
	// nodes being known and all nodes having the same versions.
	{
		isUpdating := status.HasUpdatingCondition()
		notUpdated := !status.HasUpdatedCondition()
		sameCount := readyReplicas != 0 && readyReplicas == desiredReplicas
		sameVersion := allNodesHaveVersion(nodes, desiredVersion, providerOperatorVersionLabel)

		if isUpdating && notUpdated && sameCount && sameVersion {
			status.Conditions = status.WithUpdatedCondition()
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", infrastructurev1alpha2.ClusterStatusConditionUpdated))
		}
	}

	// Check all node versions held by the cluster status and add the version the
	// tenant cluster successfully migrated to, to the historical list of versions.
	{
		hasTransitioned := status.HasCreatedCondition() || status.HasUpdatedCondition()
		notSet := !status.HasVersion(desiredVersion)
		sameCount := readyReplicas != 0 && readyReplicas == desiredReplicas
		sameVersion := allNodesHaveVersion(nodes, desiredVersion, providerOperatorVersionLabel)

		if hasTransitioned && notSet && sameCount && sameVersion {
			status.Versions = status.WithNewVersion(desiredVersion)
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting status versions with new version: %q", desiredVersion))
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
