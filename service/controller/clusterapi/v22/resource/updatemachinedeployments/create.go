package updatemachinedeployments

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v21/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var machineDeployments []v1alpha1.MachineDeployment
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

	for _, md := range machineDeployments {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating machine deployment %#q for tenant cluster %#q", md.Namespace+"/"+md.Name, key.ClusterID(&cr)))

		var updated bool

		// Syncing the Provider Operator version. For instance aws-operator,
		// kvm-operator or the like.
		{
			l := fmt.Sprintf("%s-operator.giantswarm.io/version", r.provider)
			d, ok := cr.Labels[l]
			c := md.Labels[l]
			if ok && d != "" && d != md.Labels[l] {
				md.Labels[l] = d
				updated = true

				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("label value of %#q changed from %#q to %#q", l, c, d))
			}
		}

		// Syncing the cluster-operator version.
		{
			l := label.OperatorVersion
			d, ok := cr.Labels[l]
			c := md.Labels[l]
			if ok && d != "" && d != md.Labels[l] {
				md.Labels[l] = d
				updated = true

				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("label value of %#q changed from %#q to %#q", l, c, d))
			}
		}

		// Syncing the Giant Swarm Release version.
		{
			l := label.ReleaseVersion
			d, ok := cr.Labels[l]
			c := md.Labels[l]
			if ok && d != "" && d != md.Labels[l] {
				md.Labels[l] = d
				updated = true

				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("label value of %#q changed from %#q to %#q", l, c, d))
			}
		}

		if updated {
			_, err := r.cmaClient.ClusterV1alpha1().MachineDeployments(md.Namespace).Update(&md)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated machine deployment %#q for tenant cluster %#q", md.Namespace+"/"+md.Name, key.ClusterID(&cr)))
	}

	return nil
}
