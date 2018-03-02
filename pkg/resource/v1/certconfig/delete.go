package certconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	return nil
}

// NewDeletePatch is called upon observed custom object deletion. It receives
// the deleted custom object, the current state as provided by GetCurrentState
// and the desired state as provided by GetDesiredState. NewDeletePatch
// analyses the current and desired state and returns the patch to be applied by
// Create, Update and Delete functions.
func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	delete, err := r.newDeleteChangeForDeletePatch(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) newDeleteChangeForDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) ([]*v1alpha1.CertConfig, error) {
	currentCertConfigs, err := toCertConfigs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "finding out which certconfigs have to be deleted")

	r.logger.LogCtx(ctx, "debug", fmt.Sprintf("found %d certconfigs that have to be deleted", len(currentCertConfigs)))

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

	r.logger.LogCtx(ctx, "debug", "finding out which certconfigs have to be deleted")

	var certConfigsToDelete []*v1alpha1.CertConfig
	for _, currentCertConfig := range currentCertConfigs {
		_, err := getCertConfigByName(desiredCertConfigs, currentCertConfig.Name)
		// Existing CertConfig is not desired anymore so it should be deleted.
		if IsNotFound(err) {
			certConfigsToDelete = append(certConfigsToDelete, currentCertConfig)
		}
	}

	r.logger.LogCtx(ctx, "debug", fmt.Sprintf("found %d certconfigs that have to be deleted", len(certConfigsToDelete)))

	return certConfigsToDelete, nil
}
