package encryptionkey

import (
	"context"

	"github.com/giantswarm/microerror"
	"k8s.io/api/core/v1"
)

// ApplyCreateChange takes observed custom object and create portion of the
// Patch provided by NewUpdatePatch or NewDeletePatch. It creates k8s secret
// for encryption key if needed.
func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	secret, err := toSecret(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if secret == nil {
		return microerror.Maskf(invalidValueError, "createChange (*v1.Secret) must not be nil")
	}

	_, err = r.k8sClient.Core().Secrets(v1.NamespaceDefault).Create(secret)
	if err != nil {
		err = microerror.Mask(err)
	}

	return err
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (*v1.Secret, error) {
	currentSecret, err := toSecret(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredSecret, err := toSecret(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "finding out if the secret has to be created")

	var secretToCreate *v1.Secret
	if currentSecret == nil {
		secretToCreate = desiredSecret
	}

	r.logger.LogCtx(ctx, "debug", "found out if the secret has to be created")

	return secretToCreate, nil
}
