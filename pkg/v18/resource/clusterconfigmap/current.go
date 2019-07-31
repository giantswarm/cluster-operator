package clusterconfigmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/v18/key"
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
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil, nil
	}

	clusterConfig, err := r.getClusterConfigFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Cluster namespace is created by the provider operator. If it doesn't
	// exist yet we should retry in the next reconciliation loop.
	ns, err := r.k8sClient.CoreV1().Namespaces().Get(clusterConfig.ID, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("cluster namespace %#q does not exist", clusterConfig.ID))
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil, nil
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("FOUND NAMESPACE %#v", ns))
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("FOUND NAMESPACE STATUS %#v", ns.Status)

	name := key.ClusterConfigMapName(clusterConfig)

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding cluster configMap %#q in namespace %#q", name, clusterConfig.ID))

	cm, err := r.k8sClient.CoreV1().ConfigMaps(clusterConfig.ID).Get(name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find cluster configMap %#q in namespace %#q", name, clusterConfig.ID))
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found cluster configMap %#q in namespace %#q", name, clusterConfig.ID))

	return []*corev1.ConfigMap{cm}, nil
}
