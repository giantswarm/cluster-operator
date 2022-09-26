package deletecrs

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/giantswarm/cluster-operator/v5/pkg/label"
	"github.com/giantswarm/cluster-operator/v5/service/controller/key"
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

	var list *unstructured.UnstructuredList
	{
		gvk, err := apiutil.GVKForObject(r.newObjFunc(), r.k8sClient.Scheme())
		if err != nil {
			return microerror.Mask(err)
		}
		gvk.Kind += "List"

		l := &unstructured.UnstructuredList{}
		l.SetGroupVersionKind(gvk)

		list = l
	}

	{
		r.logger.Debugf(ctx, "finding objects of type %T for tenant cluster %#q", r.newObjFunc(), key.ClusterID(cr))

		err = r.k8sClient.CtrlClient().List(
			ctx,
			list,
			client.InNamespace(cr.GetNamespace()),
			client.MatchingLabels{label.Cluster: key.ClusterID(cr)},
		)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "found %d object(s) of type %T for tenant cluster %#q", len(list.Items), r.newObjFunc(), key.ClusterID(cr))
	}

	for _, i := range list.Items {
		i := i // dereferencing pointer value into new scope

		r.logger.Debugf(ctx, "deleting object %#q of type %T for tenant cluster %#q", fmt.Sprintf("%s/%s", i.GetNamespace(), i.GetName()), r.newObjFunc(), key.ClusterID(cr))

		err = r.k8sClient.CtrlClient().Delete(ctx, &i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "deleted object %#q of type %T for tenant cluster %#q", fmt.Sprintf("%s/%s", i.GetNamespace(), i.GetName()), r.newObjFunc(), key.ClusterID(cr))
	}

	return nil
}
