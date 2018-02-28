package certconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)

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

	r.logger.LogCtx(ctx, "debug", "finding out which certconfigs have to be updated")

	var certConfigsToUpdate []*v1alpha1.CertConfig
	for _, currentCertConfig := range currentCertConfigs {
		desiredCertConfig, err := getCertConfigByName(desiredCertConfigs, currentCertConfig.Name)
		if IsNotFound(err) {
			// NOTE that this case indicates we should remove the current
			// certconfig eventually.
			r.logger.LogCtx(ctx, "warning", fmt.Sprintf("not updating certconfig '%s': no desired certconfig found", currentCertConfig.GetName()))
			continue
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		if isCertConfigModified(desiredCertConfig, currentCertConfig) {
			r.logger.LogCtx(ctx, "debug", fmt.Sprintf("found certconfig '%s' that has to be updated", desiredCertConfig.GetName()))
			certConfigsToUpdate = append(certConfigsToUpdate, desiredCertConfig)
		} else {
			r.logger.LogCtx(ctx, "debug", fmt.Sprintf("not updating certconfig '%s': no changes found", currentCertConfig.GetName()))
		}
	}

	return certConfigsToUpdate, nil
}
