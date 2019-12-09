package certconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/resource/crud"
)

// ApplyUpdateChange takes observed custom object and update portion of the
// Patch provided by NewUpdatePatch or NewDeletePatch.
func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	certConfigs, err := toCertConfigs(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(certConfigs) > 0 {
		for _, certConfig := range certConfigs {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating certconfig %#q in namespace %#q", certConfig.Name, certConfig.Namespace))

			_, err = r.g8sClient.CoreV1alpha1().CertConfigs(certConfig.Namespace).Update(certConfig)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated certconfig %#q in namespace %#q", certConfig.Name, certConfig.Namespace))
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not update certconfigs")
	}

	return nil
}

// NewUpdatePatch computes appropriate Patch based on difference in current
// state and desired state.
func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*crud.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	delete, err := r.newDeleteChangeForUpdatePatch(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := crud.NewPatch()
	patch.SetCreateChange(create)
	patch.SetDeleteChange(delete)
	patch.SetUpdateChange(update)

	return patch, nil
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

	var certConfigsToDelete []*v1alpha1.CertConfig
	for _, currentCertConfig := range currentCertConfigs {
		_, err := getCertConfigByName(desiredCertConfigs, currentCertConfig.Name)
		if IsNotFound(err) {
			certConfigsToDelete = append(certConfigsToDelete, currentCertConfig)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return certConfigsToDelete, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) ([]*v1alpha1.CertConfig, error) {
	currentCertConfigs, err := toCertConfigs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredCertConfigs, err := toCertConfigs(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var certConfigsToUpdate []*v1alpha1.CertConfig
	for _, currentCertConfig := range currentCertConfigs {
		desiredCertConfig, err := getCertConfigByName(desiredCertConfigs, currentCertConfig.Name)
		if IsNotFound(err) {
			// Ignore here. These are handled by newDeleteChangeForUpdatePatch().
			continue
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		if isCertConfigModified(desiredCertConfig, currentCertConfig) {
			// Create a copy and set the resource version to allow the CR to be updated.
			certConfigToUpdate := desiredCertConfig.DeepCopy()
			certConfigToUpdate.ObjectMeta.ResourceVersion = currentCertConfig.ObjectMeta.ResourceVersion

			certConfigsToUpdate = append(certConfigsToUpdate, certConfigToUpdate)
		}
	}

	return certConfigsToUpdate, nil
}
