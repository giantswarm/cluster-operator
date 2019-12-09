package cleanupmachinedeployments

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/finalizerskeptcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	clusterv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var machineDeployments []clusterv1alpha2.MachineDeployment
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding machine deployments for tenant cluster")

		l := metav1.AddLabelToSelector(
			&metav1.LabelSelector{},
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

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d machine deployments for tenant cluster", len(machineDeployments)))
	}

	// We do not want to delete the Cluster CR as long as there are any
	// MachineDeployment CRs. This is because there cannot be any Node Pool
	// without a Cluster.
	if len(machineDeployments) != 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "keeping finalizers")
		finalizerskeptcontext.SetKept(ctx)
	}

	for _, md := range machineDeployments {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting machine deployment %#q for tenant cluster %#q", md.Namespace+"/"+md.Name, key.ClusterID(&cr)))

		err = r.cmaClient.ClusterV1alpha1().MachineDeployments(md.Namespace).Delete(md.Name, &metav1.DeleteOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted machine deployment %#q for tenant cluster %#q", md.Namespace+"/"+md.Name, key.ClusterID(&cr)))
	}

	return nil
}
