package workercount

import (
	"context"
	"fmt"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/v22/controllercontext"
	"github.com/giantswarm/cluster-operator/pkg/v22/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := r.toClusterConfigFunc(obj)
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

	var workerCount int
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding nodes of tenant cluster %#q", key.ClusterID(cr)))

		o := metav1.ListOptions{
			// This label selector excludes the master nodes from node list.
			//
			// Constructing this LabelSelector is not currently possible with
			// k8s types and functions. Therefore it's hardcoded here.
			LabelSelector: fmt.Sprintf("!%s", label.MasterNodeRole),
		}

		l, err := cc.Client.TenantCluster.K8s.CoreV1().Nodes().List(o)
		if tenant.IsAPINotAvailable(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant API not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		workerCount = len(l.Items)

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d nodes of tenant cluster %#q", len(l.Items), key.ClusterID(cr)))
	}

	{
		cc.Status.Worker.Nodes = int32(workerCount)
	}

	return nil
}
