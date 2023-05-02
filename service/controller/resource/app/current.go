package app

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v8/pkg/controller/context/resourcecanceledcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/v5/pkg/label"
	"github.com/giantswarm/cluster-operator/v5/pkg/project"
	"github.com/giantswarm/cluster-operator/v5/service/controller/key"
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
			LabelSelector: fmt.Sprintf("%s=%s,%s=%s", label.Cluster, key.ClusterID(&cr), label.ManagedBy, project.Name()),
		}

		list := &v1alpha1.AppList{}

		err := r.ctrlClient.List(ctx, list, &client.ListOptions{Namespace: key.ClusterID(&cr), Raw: &o})
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
