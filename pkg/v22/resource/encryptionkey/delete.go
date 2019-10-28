package encryptionkey

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ApplyDeleteChange takes observed custom object and delete portion of the
// Patch provided by NewUpdatePatch and NewDeletePatch. It deletes k8s secret
// for related encryption key if needed.
func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	objectMeta, err := r.toClusterObjectMetaFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	secret, err := toSecret(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "deleting encryptionkey secret")

	if secret != nil {
		err = r.k8sClient.CoreV1().Secrets(objectMeta.Namespace).Delete(secret.Name, &metav1.DeleteOptions{})
		if apierrors.IsNotFound(err) {
			// It's ok if secret doesn't exist anymore. It would have been
			// deleted anyway. Rational reason for this is during migration
			// period from kubernetesd to cluster-operator when there's a race
			// between these on which one is first to delete pending resource.
			err = nil
		} else if err != nil {
			err = microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting encryptionkey secret: deleted")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting encryptionkey secret: already deleted")
	}

	return err
}

// NewDeletePatch is called upon observed custom object deletion. It receives
// the deleted custom object, the current state as provided by GetCurrentState
// and the desired state as provided by GetDesiredState. NewDeletePatch
// analyses the current and desired state and returns the patch to be applied by
// Create, Delete, and Update functions.
func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	return nil, nil
}
