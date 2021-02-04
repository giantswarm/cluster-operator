package appversionlabel

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/v3/pkg/project"
	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var apps []*v1alpha1.App
	{
		r.logger.Debugf(ctx, "finding optional apps for tenant cluster %#q", key.ClusterID(&cr))

		o := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s!=%s", label.ManagedBy, project.Name()),
		}
		list, err := r.g8sClient.ApplicationV1alpha1().Apps(key.ClusterID(&cr)).List(ctx, o)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, item := range list.Items {
			apps = append(apps, item.DeepCopy())
		}

		r.logger.Debugf(ctx, "found %d optional apps for tenant cluster %#q", len(apps), key.ClusterID(&cr))
	}

	for _, app := range apps {
		r.logger.Debugf(ctx, "deleting App CR %#q in namespace %#q", app.Name, app.Namespace)

		err := r.g8sClient.ApplicationV1alpha1().Apps(app.Namespace).Delete(ctx, app.Name, metav1.DeleteOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.Debugf(ctx, "already deleted app %#q in namespace %#q", app.Name, app.Namespace)
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			r.logger.Debugf(ctx, "deleted app %#q in namespace %#q", app.Name, app.Namespace)
		}

	}

	return nil
}
