package clusterconfigmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/pkg/v21/key"
)

func (r *StateGetter) GetCurrentState(ctx context.Context, obj interface{}) ([]*corev1.ConfigMap, error) {
	objectMeta, err := r.getClusterObjectMetaFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Cluster configMap is deleted by the provider operator when it deletes
	// the tenant cluster namespace in the control plane cluster.
	if key.IsDeleted(objectMeta) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "redirecting cluster configMap deletion to provider operators")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)

		return nil, nil
	}

	clusterConfig, err := r.getClusterConfigFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Cluster namespace is created by the provider operator. If it doesn't
	// exist yet we should retry in the next reconciliation loop.
	_, err = r.k8sClient.CoreV1().Namespaces().Get(key.ClusterID(clusterConfig), metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("cluster namespace %#q does not exist", key.ClusterID(clusterConfig)))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)

		return nil, nil
	}

	var configMaps []*corev1.ConfigMap
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding cluster configMaps in namespace %#q", key.ClusterID(clusterConfig)))

		lo := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", label.ManagedBy, project.Name()),
		}

		list, err := r.k8sClient.CoreV1().ConfigMaps(clusterConfig.ID).List(lo)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		for _, item := range list.Items {
			configMaps = append(configMaps, item.DeepCopy())
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d configMaps in namespace %#q", len(configMaps), key.ClusterID(clusterConfig)))
	}

	return configMaps, nil
}
