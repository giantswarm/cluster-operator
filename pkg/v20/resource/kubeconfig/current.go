package kubeconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/v20/key"
)

func (r *StateGetter) GetCurrentState(ctx context.Context, obj interface{}) ([]*corev1.Secret, error) {
	objectMeta, err := r.getClusterObjectMetaFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Cluster configMap is deleted by the provider operator when it deletes
	// the tenant cluster namespace in the control plane cluster.
	if key.IsDeleted(objectMeta) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "redirecting kubeconfig secret deletion to provider operators")
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil, nil
	}

	clusterConfig, err := r.getClusterConfigFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	secretName := key.KubeConfigSecretName(clusterConfig)

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding kubeconfig secret %#q", secretName))

	secret, err := r.k8sClient.CoreV1().Secrets(clusterConfig.ID).Get(secretName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find kubeconfig secret %#q", secretName))
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found kubeconfig secret %#q", secretName))

	return []*corev1.Secret{secret}, nil
}
