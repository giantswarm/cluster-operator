package cleanupmachinedeployments

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/finalizerskeptcontext"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	list := &apiv1alpha2.MachineDeploymentList{}
	{
		err = r.k8sClient.CtrlClient().List(
			ctx,
			list,
			client.InNamespace(cr.Namespace),
			client.MatchingLabels{label.Cluster: key.ClusterID(&cr)},
		)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// We do not want to delete the Cluster CR as long as there are any
	// MachineDeployment CRs. This is because there cannot be any Node Pool
	// without a Cluster.
	if len(list.Items) != 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d objects of type %T for tenant cluster %#q", len(list.Items), list, key.ClusterID(&cr)))
		r.logger.LogCtx(ctx, "level", "debug", "message", "keeping finalizers")
		finalizerskeptcontext.SetKept(ctx)
	}

	return nil
}
