package updateinfrarefs

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	or, err := r.toObjRef(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(cc.Status.Versions) == 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no release versions in controller context yet")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	// Here we fetch the provider specific CR defined as infrastructure reference
	// in the CAPI type. We use an unstructured object and therefore need to set
	// the api version and kind accordingly. If we would not do that the
	// controller-runtime client cannot find the right object.
	ir := &unstructured.Unstructured{}
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding infrastructure reference")

		ir.SetAPIVersion(or.APIVersion)
		ir.SetKind(or.Kind)

		err = r.k8sClient.CtrlClient().Get(ctx, key.ObjRefToNamespacedName(or), ir)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found infrastructure reference")
	}

	var updated bool

	// Syncing the provider operator version label, e.g. for aws-operator,
	// kvm-operator or the like.
	{
		l := fmt.Sprintf("%s-operator.giantswarm.io/version", r.provider)
		d := cc.Status.Versions[l]
		c, ok := ir.GetLabels()[l]
		if ok && d != "" && d != c {
			labels := ir.GetLabels()
			labels[l] = d
			ir.SetLabels(labels)
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
			labels := ir.GetLabels()
			labels[l] = d
			ir.SetLabels(labels)
			updated = true

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("label value of %#q changed from %#q to %#q", l, c, d))
		}
	}

	if updated {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating infrastructure reference %#q", ir.GetNamespace()+"/"+ir.GetName()))

		err = r.k8sClient.CtrlClient().Update(ctx, ir)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated infrastructure reference %#q", ir.GetNamespace()+"/"+ir.GetName()))
	}

	return nil
}
