package certconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	certConfigs, err := toCertConfigs(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(certConfigs) != 0 {
		for _, certConfig := range certConfigs {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting certconfig %#q in namespace %#q", certConfig.Name, certConfig.Namespace))

			err := r.g8sClient.CoreV1alpha1().CertConfigs(certConfig.Namespace).Delete(certConfig.Name, &metav1.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted certconfig %#q in namespace %#q", certConfig.Name, certConfig.Namespace))
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not delete certconfigs")
	}

	return nil
}

// NewDeletePatch is called upon observed custom object deletion. It receives
// the deleted custom object, the current state as provided by GetCurrentState
// and the desired state as provided by GetDesiredState. NewDeletePatch
// analyses the current and desired state and returns the patch to be applied by
// Create, Update and Delete functions.
func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	delete, err := r.newDeleteChangeForDeletePatch(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) newDeleteChangeForDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) ([]*v1alpha1.CertConfig, error) {
	currentCertConfigs, err := toCertConfigs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return currentCertConfigs, nil
}
