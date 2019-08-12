package encryptionkey

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v18/key"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	secret, err := toSecret(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if secret != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting secret %#q", key.EncryptionKeySecretName(cr)))

		err = r.k8sClient.Core().Secrets(secret.Namespace).Delete(secret.Name, &metav1.DeleteOptions{})
		if apierrors.IsNotFound(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted secret %#q", key.EncryptionKeySecretName(cr)))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not delete secret %#q", key.EncryptionKeySecretName(cr)))
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	// Upon deletion we only remove what we found in the system. That should
	// cleanup the resource we manage properly.
	return currentState, nil
}
