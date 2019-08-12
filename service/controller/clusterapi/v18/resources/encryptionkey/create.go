package encryptionkey

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v18/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	secret, err := toSecret(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if secret != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating secret %#q", key.EncryptionKeySecretName(cr)))

		_, err = r.k8sClient.Core().Secrets(secret.Namespace).Create(secret)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created secret %#q", key.EncryptionKeySecretName(cr)))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not create secret %#q", key.EncryptionKeySecretName(cr)))
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (*corev1.Secret, error) {
	currentSecret, err := toSecret(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredSecret, err := toSecret(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var secretToCreate *corev1.Secret
	if currentSecret == nil {
		secretToCreate = desiredSecret
	}

	return secretToCreate, nil
}
