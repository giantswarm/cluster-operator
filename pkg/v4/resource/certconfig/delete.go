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
	objectMeta, err := r.toClusterObjectMetaFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	certConfigsToDelete, err := toCertConfigs(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(certConfigsToDelete) != 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting the certconfigs in the Kubernetes API")

		for _, certConfig := range certConfigsToDelete {
			err := r.g8sClient.CoreV1alpha1().CertConfigs(objectMeta.Namespace).Delete(certConfig.Name, &metav1.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deleted the certconfigs in the Kubernetes API")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the certconfigs do not need to be deleted from the Kubernetes API")
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

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out which certconfigs have to be deleted")

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d certconfigs that have to be deleted", len(currentCertConfigs)))

	return currentCertConfigs, nil
}

func (r *Resource) newDeleteChangeForUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) ([]*v1alpha1.CertConfig, error) {
	currentCertConfigs, err := toCertConfigs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredCertConfigs, err := toCertConfigs(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out which certconfigs have to be deleted")

	var certConfigsToDelete []*v1alpha1.CertConfig
	for _, currentCertConfig := range currentCertConfigs {
		_, err := getCertConfigByName(desiredCertConfigs, currentCertConfig.Name)
		// Existing CertConfig is not desired anymore so it should be deleted.
		if IsNotFound(err) {
			certConfigsToDelete = append(certConfigsToDelete, currentCertConfig)
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d certconfigs that have to be deleted", len(certConfigsToDelete)))

	return certConfigsToDelete, nil
}
