package updateg8scontrolplanes

import (
	"context"
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	cpList := &infrastructurev1alpha2.G8sControlPlaneList{}
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding G8sControlPlanes for tenant cluster")

		err = r.k8sClient.CtrlClient().List(
			ctx,
			cpList,
			client.InNamespace(cr.Namespace),
			client.MatchingLabels{label.Cluster: key.ClusterID(&cr)},
		)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d G8sControlPlanes for tenant cluster", len(cpList.Items)))
	}

	for _, cp := range cpList.Items {
		cp := cp // dereferencing pointer value into new scope

		var updated bool

		// Syncing the cluster-operator version.
		{
			l := label.OperatorVersion
			d, ok := cr.Labels[l]
			c := cp.Labels[l]
			if ok && d != "" && d != c {
				cp.Labels[l] = d
				updated = true

				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("label value of %#q changed from %#q to %#q", l, c, d))
			}
		}

		// Syncing the Giant Swarm Release version.
		{
			l := label.ReleaseVersion
			d, ok := cr.Labels[l]
			c := cp.Labels[l]
			if ok && d != "" && d != c {
				cp.Labels[l] = d
				updated = true

				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("label value of %#q changed from %#q to %#q", l, c, d))
			}
		}

		if updated {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating G8sControlPlane %#q for tenant cluster %#q", cp.Namespace+"/"+cp.Name, key.ClusterID(&cr)))

			err = r.k8sClient.CtrlClient().Update(ctx, &cp)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated G8sControlPlane %#q for tenant cluster %#q", cp.Namespace+"/"+cp.Name, key.ClusterID(&cr)))
		}
	}

	return nil
}
