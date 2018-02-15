package encryptionkey

import (
	"context"

	"github.com/giantswarm/microerror"
)

// ApplyCreateChange takes observed custom object and create portion of the
// Patch provided by NewUpdatePatch or NewDeletePatch. It creates k8s secret
// for encryption key if needed.
func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentSecret, err := toSecret(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredSecret, err := toSecret(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "finding out if the secret has to be created")

	// secretToCreate must be an empty interface instead of *v1.Secret because
	// that makes a difference when comparing return value for nil.
	var secretToCreate interface{}
	if currentSecret == nil {
		secretToCreate = desiredSecret
	}

	r.logger.LogCtx(ctx, "debug", "found out if the secret has to be created")

	return secretToCreate, nil

}
