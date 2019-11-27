package workercount

import (
	"context"
	"fmt"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v22/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v22/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := r.toClusterFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		if cc.Client.TenantCluster.K8s == nil {
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster clients not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}
	}

	var nodes []corev1.Node
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding nodes of tenant cluster %#q", key.ClusterID(&cr)))

		o := metav1.ListOptions{
			// This label selector excludes the master nodes from node list.
			//
			// Constructing this LabelSelector is not currently possible with
			// k8s types and functions. Therefore it's hardcoded here.
			LabelSelector: fmt.Sprintf("!%s", label.MasterNode),
		}

		l, err := cc.Client.TenantCluster.K8s.CoreV1().Nodes().List(o)
		if tenant.IsAPINotAvailable(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant API not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		nodes = l.Items

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found nodes of tenant cluster %#q", key.ClusterID(&cr)))
	}

	{

		for _, n := range nodes {
			id := n.Labels[label.MachineDeployment]

			if cc.Status.Worker == nil {
				cc.Status.Worker = map[string]controllercontext.ContextStatusWorker{}
			}

			{
				val := cc.Status.Worker[id]
				val.Nodes++
				cc.Status.Worker[id] = val
			}

			for _, c := range n.Status.Conditions {
				if c.Type == corev1.NodeReady && c.Status == corev1.ConditionTrue {
					val := cc.Status.Worker[id]
					val.Ready++
					cc.Status.Worker[id] = val
				}
			}
		}
	}

	return nil
}
