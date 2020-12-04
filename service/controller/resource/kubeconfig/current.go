package kubeconfig

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) ([]*corev1.Secret, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// The secrets are deleted when the namespace is deleted.
	if key.IsDeleted(&cr) {
		r.logger.Debugf(ctx, "not deleting kubeconfig secret for tenant cluster %#q", key.ClusterID(&cr))
		r.logger.Debugf(ctx, "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil
	}

	var secret *corev1.Secret
	{
		r.logger.Debugf(ctx, "finding secret %#q for tenant cluster %#q", key.KubeConfigSecretName(&cr), key.ClusterID(&cr))

		secret, err = r.k8sClient.CoreV1().Secrets(key.ClusterID(&cr)).Get(ctx, key.KubeConfigSecretName(&cr), metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.Debugf(ctx, "did not find secret %#q for tenant cluster %#q", key.KubeConfigSecretName(&cr), key.ClusterID(&cr))
			return nil, nil

		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "found secret %#q for tenant cluster %#q", key.KubeConfigSecretName(&cr), key.ClusterID(&cr))
	}

	return []*corev1.Secret{secret}, nil
}
