package machinedeploymentstatus

import (
	"context"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"github.com/giantswarm/tenantcluster"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v17/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v17/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToMachineDeployment(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var nodes []corev1.Node
	var ready []corev1.Node
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding nodes of tenant cluster")

		l, err := cc.Client.TenantCluster.K8s.CoreV1().Nodes().List(metav1.ListOptions{})
		if tenantcluster.IsTimeout(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "timeout fetching certificates")

			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)

			return nil
		} else if tenant.IsAPINotAvailable(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant API not available")

			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		nodes = l.Items

		for _, n := range nodes {
			for _, c := range n.Status.Conditions {
				if c.Type == corev1.NodeReady && c.Status == corev1.ConditionTrue {
					ready = append(ready, n)
				}
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found nodes of tenant cluster")

		replicasChanged := cr.Status.Replicas != int32(len(nodes))
		readyReplicasChanged := cr.Status.ReadyReplicas != int32(len(ready))

		if !replicasChanged && !readyReplicasChanged {
			r.logger.LogCtx(ctx, "level", "debug", "message", "not updating status of machine deployment")
			return nil
		}
	}

	var resourceVersion string
	{
		m, err := r.cmaClient.ClusterV1alpha1().MachineDeployments(cr.Namespace).Get(cr.Name, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		resourceVersion = m.ResourceVersion
	}

	{
		cr.ResourceVersion = resourceVersion
		cr.Status.Replicas = int32(len(nodes))
		cr.Status.ReadyReplicas = int32(len(ready))
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating status of machine deployment")

		_, err := r.cmaClient.ClusterV1alpha1().MachineDeployments(cr.Namespace).Update(&cr)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated status of machine deployment")

		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
	}

	return nil
}
