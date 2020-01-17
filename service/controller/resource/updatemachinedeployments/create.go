package updatemachinedeployments

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	mdList := &apiv1alpha2.MachineDeploymentList{}
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding MachineDeployments for tenant cluster")

		err = r.k8sClient.CtrlClient().List(
			ctx,
			mdList,
			client.InNamespace(cr.Namespace),
			client.MatchingLabels{label.Cluster: key.ClusterID(&cr)},
		)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d MachineDeployments for tenant cluster", len(mdList.Items)))
	}

	for _, md := range mdList.Items {
		var updated bool

		// Syncing the cluster-operator version.
		{
			l := label.OperatorVersion
			d, ok := cr.Labels[l]
			c := md.Labels[l]
			if ok && d != "" && d != c {
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
			if ok && d != "" && d != c {
				md.Labels[l] = d
				updated = true

				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("label value of %#q changed from %#q to %#q", l, c, d))
			}
		}

		if updated {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating machine deployment %#q for tenant cluster %#q", md.Namespace+"/"+md.Name, key.ClusterID(&cr)))

			err = r.k8sClient.CtrlClient().Update(ctx, &md)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated machine deployment %#q for tenant cluster %#q", md.Namespace+"/"+md.Name, key.ClusterID(&cr)))
		}
	}

	return nil
}
