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
	objectMeta, err := r.toClusterObjectMetaFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	certConfigsToUpdate, err := toCertConfigs(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(certConfigsToUpdate) > 0 {
		_ = r.logger.LogCtx(ctx, "level", "debug", "message", "updating certconfigs")

		for _, certConfigToUpdate := range certConfigsToUpdate {
			_, err = r.g8sClient.CoreV1alpha1().CertConfigs(objectMeta.Namespace).Update(certConfigToUpdate)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		_ = r.logger.LogCtx(ctx, "level", "debug", "message", "updated certconfigs")
	} else {
		_ = r.logger.LogCtx(ctx, "level", "debug", "message", "no need to update certconfigs")
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

	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	delete, err := r.newDeleteChangeForUpdatePatch(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := crud.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)
	patch.SetDeleteChange(delete)

	return patch, nil
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

	_ = r.logger.LogCtx(ctx, "level", "debug", "message", "finding out which certconfigs have to be updated")

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
			_ = r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found certconfig '%s' that has to be updated", desiredCertConfig.GetName()))

			// Create a copy and set the resource version to allow the CR to be updated.
			certConfigToUpdate := desiredCertConfig.DeepCopy()
			certConfigToUpdate.ObjectMeta.ResourceVersion = currentCertConfig.ObjectMeta.ResourceVersion

			certConfigsToUpdate = append(certConfigsToUpdate, certConfigToUpdate)
		} else {
			_ = r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not updating certconfig '%s': no changes found", currentCertConfig.GetName()))
		}
	}

	return certConfigsToUpdate, nil
}
