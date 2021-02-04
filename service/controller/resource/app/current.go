package app

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/v3/pkg/project"
	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) ([]*v1alpha1.App, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var labelSelector string
	{
		// On deletion we delete all app CRs except app-operator. It is deleted
		// with the cluster namespace.
		if key.IsDeleted(&cr) {
			labelSelector = fmt.Sprintf("%s=%s,%s!=%s", label.ManagedBy, project.Name(), label.AppKubernetesName, "app-operator")
		} else {
			labelSelector = fmt.Sprintf("%s=%s", label.ManagedBy, project.Name())
		}
	}

	var apps []*v1alpha1.App
	{
		r.logger.Debugf(ctx, "finding apps for tenant cluster %#q", key.ClusterID(&cr))

		o := metav1.ListOptions{
			LabelSelector: labelSelector,
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
