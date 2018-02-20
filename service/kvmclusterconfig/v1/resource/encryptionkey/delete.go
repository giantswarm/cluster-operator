package encryptionkey

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ApplyDeleteChange takes observed custom object and delete portion of the
// Patch provided by NewUpdatePatch and NewDeletePatch. It deletes k8s secret
// for related encryption key if needed.
func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	secret, err := toSecret(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "deleting encryptionkey secret")

	if secret != nil {
		err = r.k8sClient.Core().Secrets(v1.NamespaceDefault).Delete(secret.Name, &metav1.DeleteOptions{})
		if err != nil {
			err = microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "debug", "deleting encryptionkey secret: deleted")
	} else {
		r.logger.LogCtx(ctx, "debug", "deleting encryptionkey secret: already deleted")
	}

	return err
}

// NewDeletePatch is called upon observed custom object deletion. It receives
// the deleted custom object, the current state as provided by GetCurrentState
// and the desired state as provided by GetDesiredState. NewDeletePatch
// analyses the current and desired state and returns the patch to be applied by
// Create, Delete, and Update functions.
func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	return nil, nil
}
