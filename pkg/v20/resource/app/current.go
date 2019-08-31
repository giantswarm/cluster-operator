package app

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/pkg/v20/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) ([]*v1alpha1.App, error) {
	objectMeta, err := r.getClusterObjectMetaFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Cluster configMap is deleted by the provider operator when it deletes
	// the tenant cluster namespace in the control plane cluster.
	if key.IsDeleted(objectMeta) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "redirecting kubeconfig secret deletion to provider operators")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)

		return nil, nil
	}

	clusterConfig, err := r.getClusterConfigFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var apps []*v1alpha1.App
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding apps in tenant cluster %#q", clusterConfig.ID))

		o := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", label.ManagedBy, project.Name()),
		}

		list, err := r.g8sClient.ApplicationV1alpha1().Apps(clusterConfig.ID).List(o)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		for _, item := range list.Items {
			apps = append(apps, item.DeepCopy())
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d apps in tenant cluster %#q", len(apps), clusterConfig.ID))
	}

	return apps, nil
}
