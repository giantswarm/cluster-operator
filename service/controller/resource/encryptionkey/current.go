package encryptionkey

import (
	"context"

	"github.com/giantswarm/microerror"
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

	var secret *corev1.Secret
	{
		r.logger.Debugf(ctx, "finding secret %#q in namespace %#q", secretName(cr), cr.Namespace)

		secret, err = r.k8sClient.CoreV1().Secrets(cr.Namespace).Get(ctx, secretName(cr), metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.Debugf(ctx, "did not find secret %#q in namespace %#q", secretName(cr), cr.Namespace)
			return nil, nil
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "found secret %#q in namespace %#q", secretName(cr), cr.Namespace)
	}

	return []*corev1.Secret{secret}, nil
}
