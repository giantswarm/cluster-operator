package app

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/resourcecanceledcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/v3/pkg/label"
	"github.com/giantswarm/cluster-operator/v3/pkg/project"
	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) ([]*v1alpha1.App, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// The app custom resources are deleted when the namespace is deleted.
	if key.IsDeleted(&cr) {
		r.logger.Debugf(ctx, "not deleting apps for tenant cluster %#q", key.ClusterID(&cr))
		r.logger.Debugf(ctx, "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil
	}

	var apps []*v1alpha1.App
	{
		r.logger.Debugf(ctx, "finding apps for tenant cluster %#q", key.ClusterID(&cr))

		o := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", label.ManagedBy, project.Name()),
		}

		list, err := r.g8sClient.ApplicationV1alpha1().Apps(key.ClusterID(&cr)).List(ctx, o)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		for _, item := range list.Items {
			apps = append(apps, item.DeepCopy())
		}

		r.logger.Debugf(ctx, "found %d apps for tenant cluster %#q", len(apps), key.ClusterID(&cr))
	}

	return apps, nil
}
