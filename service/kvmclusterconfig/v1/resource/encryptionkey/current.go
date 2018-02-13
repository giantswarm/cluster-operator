package encryptionkey

import (
	"context"

	"github.com/giantswarm/microerror"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/service/kvmclusterconfig/v1/key"
)

// GetCurrentState takes observed custom object as an input and based on that
// information looks for current state of cluster encryption key secret and
// returns it. Return value is of type *v1.Secret.
func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	secretName := key.EncryptionKeySecretName(customObject)

	r.logger.LogCtx(ctx, "debug", "looking for encryptionkey secret in the Kubernetes API", "secretName", secretName)

	secret, err := r.k8sClient.Core().Secrets(v1.NamespaceDefault).Get(secretName, apismetav1.GetOptions{})

	if apierrors.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	return secret, nil
}
