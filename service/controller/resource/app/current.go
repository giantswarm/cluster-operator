package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) ([]*v1alpha1.App, error) {
	objectMeta, err := r.getClusterObjectMetaFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Apps are deleted by the provider operator when it deletes
	// the tenant cluster namespace in the control plane cluster.
	if key.IsDeleted(objectMeta) {
		_ = r.logger.LogCtx(ctx, "level", "debug", "message", "redirecting app deletion to provider operators")
		_ = r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)

		return nil, nil
	}

	clusterConfig, err := r.getClusterConfigFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Cluster namespace is created by the provider operator. If it doesn't
	// exist yet we should retry in the next reconciliation loop.
	_, err = r.k8sClient.CoreV1().Namespaces().Get(clusterConfig.ID, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_ = r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("cluster namespace %#q does not exist", clusterConfig.ID))
		_ = r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)

		return nil, nil
	}

	var apps []*v1alpha1.App
	{
		_ = r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding apps in tenant cluster %#q", clusterConfig.ID))

		selectorLabels := []string{
			fmt.Sprintf("%s=%s", label.ManagedBy, project.Name()),
		}

		if r.provider == "azure" || r.provider == "kvm" {
			selectorLabels = append(selectorLabels, "app!=nginx-ingress-controller")
		}

		o := metav1.ListOptions{
			LabelSelector: strings.Join(selectorLabels, ","),
		}

		list, err := r.g8sClient.ApplicationV1alpha1().Apps(clusterConfig.ID).List(o)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		for _, item := range list.Items {
			apps = append(apps, item.DeepCopy())
		}

		_ = r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d apps in tenant cluster %#q", len(apps), clusterConfig.ID))
	}

	return apps, nil
}
