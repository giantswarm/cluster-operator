package clusterconfigmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) ([]*corev1.ConfigMap, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// The cluster config map is deleted implicitly by the provider operator when
	// it deletes the tenant cluster namespace in the control plane.
	if key.IsDeleted(&cr) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not deleting config map %#q for tenant cluster %#q", key.ClusterConfigMapName(&cr), key.ClusterID(&cr)))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil
	}

	var configMap *corev1.ConfigMap
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding config map %#q for tenant cluster %#q", key.ClusterConfigMapName(&cr), key.ClusterID(&cr)))

		configMap, err = r.k8sClient.CoreV1().ConfigMaps(key.ClusterID(&cr)).Get(key.ClusterConfigMapName(&cr), metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find config map %#q for tenant cluster %#q", key.ClusterConfigMapName(&cr), key.ClusterID(&cr)))
			return nil, nil
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found config map %#q for tenant cluster %#q", key.ClusterConfigMapName(&cr), key.ClusterID(&cr)))
	}

	return []*corev1.ConfigMap{configMap}, nil
}
