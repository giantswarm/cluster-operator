package certconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// ApplyCreateChange takes observed custom object and create portion of the
// Patch provided by NewUpdatePatch or NewDeletePatch. It creates CertConfig
// objects when new cluster certificates are needed.
func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	certConfigsToCreate, err := toCertConfigs(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(certConfigsToCreate) > 0 {
		r.logger.LogCtx(ctx, "debug", "creating certconfigs")

		for _, certConfigToCreate := range certConfigsToCreate {
			_, err = r.g8sClient.CoreV1alpha1().CertConfigs(v1.NamespaceDefault).Create(certConfigToCreate)
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "debug", "created certconfigs")
	} else {
		r.logger.LogCtx(ctx, "debug", "no need to create certconfigs")
	}

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
