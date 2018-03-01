package certconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) ([]*v1alpha1.CertConfig, error) {
	currentCertConfigs, err := toCertConfigs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredCertConfigs, err := toCertConfigs(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "finding out which certconfigs have to be created")

	var certConfigsToCreate []*v1alpha1.CertConfig

	for _, desiredCertConfig := range desiredCertConfigs {
		if !containsCertConfig(currentCertConfigs, desiredCertConfig) {
			certConfigsToCreate = append(certConfigsToCreate, desiredCertConfig)
		}
	}

	r.logger.LogCtx(ctx, "debug", fmt.Sprintf("found %d certconfigs that have to be created", len(certConfigsToCreate)))

	return certConfigsToCreate, nil
}
