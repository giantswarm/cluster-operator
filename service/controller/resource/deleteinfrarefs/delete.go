package deleteinfrarefs

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/giantswarm/cluster-operator/v4/pkg/label"
	"github.com/giantswarm/cluster-operator/v4/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	if r.provider != label.ProviderAWS {
		r.logger.Debugf(ctx, "provider is %q, only supported provider for %q resource is aws", r.provider, r.Name())
		r.logger.Debugf(ctx, "canceling resource")
		return nil
	}

	cr, err := meta.Accessor(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	or, err := r.toObjRef(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var ir *unstructured.Unstructured
	{
		r.logger.Debugf(ctx, "finding infrastructure reference")

		ir = &unstructured.Unstructured{}
		ir.SetAPIVersion(or.APIVersion)
		ir.SetKind(or.Kind)

		err = r.k8sClient.CtrlClient().Get(ctx, key.ObjRefToNamespacedName(or), ir)
		if apierrors.IsNotFound(err) {
			// At this point the runtime object linked in the infrastructure reference
			// does not exist anymore, which means the deletion of the parent can
			// continue now.
			r.logger.Debugf(ctx, "did not find infrastructure reference")
			r.logger.Debugf(ctx, "canceling resource")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "found infrastructure reference")
	}

	{
		r.logger.Debugf(ctx, "deleting object %#q of type %T for tenant cluster %#q", fmt.Sprintf("%s/%s", or.Namespace, or.Name), or.Kind, key.ClusterID(cr))

		err = r.k8sClient.CtrlClient().Delete(ctx, ir)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "deleted object %#q of type %T for tenant cluster %#q", fmt.Sprintf("%s/%s", or.Namespace, or.Name), or.Kind, key.ClusterID(cr))
	}

	return nil
}
