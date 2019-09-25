package encryptionkey

import (
	"context"

	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/v14patch1/key"
)

// GetCurrentState takes observed custom object as an input and based on that
// information looks for current state of cluster encryption key secret and
// returns it. Return value is of type *v1.Secret.
func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	clusterGuestConfig, err := r.toClusterGuestConfigFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	objectMeta, err := r.toClusterObjectMetaFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	secretName := key.EncryptionKeySecretName(clusterGuestConfig)

	r.logger.LogCtx(ctx, "level", "debug", "message", "looking for encryptionkey secret in the Kubernetes API", "secretName", secretName)

	secret, err := r.k8sClient.CoreV1().Secrets(objectMeta.Namespace).Get(secretName, apismetav1.GetOptions{})

	if apierrors.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not find a secret for encryptionkey in the Kubernetes API", "secretName", secretName)
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "found a secret for encryptionkey in the Kubernetes API", "secretName", secretName)

	return secret, nil
}
