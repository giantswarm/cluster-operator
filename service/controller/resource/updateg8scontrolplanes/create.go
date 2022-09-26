package updateg8scontrolplanes

import (
	"context"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/v5/pkg/label"
	"github.com/giantswarm/cluster-operator/v5/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	if r.provider != label.ProviderAWS {
		r.logger.Debugf(ctx, "provider is %q, only supported provider for %q resource is aws", r.provider, r.Name())
		r.logger.Debugf(ctx, "canceling resource")
		return nil
	}

	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	cpList := &infrastructurev1alpha3.G8sControlPlaneList{}
	{
		r.logger.Debugf(ctx, "finding G8sControlPlanes for tenant cluster")

		err = r.k8sClient.CtrlClient().List(
			ctx,
			cpList,
			client.InNamespace(cr.Namespace),
			client.MatchingLabels{label.Cluster: key.ClusterID(&cr)},
		)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "found %d G8sControlPlanes for tenant cluster", len(cpList.Items))
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

				r.logger.Debugf(ctx, "label value of %#q changed from %#q to %#q", l, c, d)
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

				r.logger.Debugf(ctx, "label value of %#q changed from %#q to %#q", l, c, d)
			}
		}

		if updated {
			r.logger.Debugf(ctx, "updating G8sControlPlane %#q for tenant cluster %#q", cp.Namespace+"/"+cp.Name, key.ClusterID(&cr))

			err = r.k8sClient.CtrlClient().Update(ctx, &cp)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.Debugf(ctx, "updated G8sControlPlane %#q for tenant cluster %#q", cp.Namespace+"/"+cp.Name, key.ClusterID(&cr))
		}
	}

	return nil
}
