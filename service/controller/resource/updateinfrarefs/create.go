package updateinfrarefs

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	nn, err := r.toNamespacedName(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	ir := &unstructured.Unstructured{}
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding infrastructure reference")

		err = r.k8sClient.CtrlClient().Get(ctx, nn, ir)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found infrastructure reference")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating infrastructure reference %#q", ir.GetNamespace()+"/"+ir.GetName()))

		var updated bool

		// Syncing the provider operator version label, e.g. for aws-operator,
		// kvm-operator or the like.
		{
			l := fmt.Sprintf("%s-operator.giantswarm.io/version", r.provider)
			d := cc.Status.Versions[l]
			c, ok := ir.GetLabels()[l]
			if ok && d != "" && d != c {
				ir.GetLabels()[l] = d
				updated = true

				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("label value of %#q changed from %#q to %#q", l, c, d))
			}
		}

		// Syncing the Giant Swarm Release version.
		{
			l := label.ReleaseVersion
			d, ok := cr.GetLabels()[l]
			c := ir.GetLabels()[l]
			if ok && d != "" && d != c {
				ir.GetLabels()[l] = d
				updated = true

				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("label value of %#q changed from %#q to %#q", l, c, d))
			}
		}

		if updated {
			err = r.k8sClient.CtrlClient().Update(ctx, ir)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated infrastructure reference %#q", ir.GetNamespace()+"/"+ir.GetName()))
	}

	return nil
}
