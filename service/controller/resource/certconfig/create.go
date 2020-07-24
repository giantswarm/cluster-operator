package certconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// applyCreateChange takes observed custom object and create portion of the
// patch provided by newUpdatePatch or newDeletePatch. It creates CertConfig
// objects when new cluster certificates are needed.
func (r *Resource) applyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	certConfigs, err := toCertConfigs(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(certConfigs) > 0 {
		for _, certConfig := range certConfigs {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating CertConfig CR %#q in namespace %#q", certConfig.Name, certConfig.Namespace))

			_, err = r.g8sClient.CoreV1alpha1().CertConfigs(certConfig.Namespace).Create(ctx, certConfig, metav1.CreateOptions{})
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created CertConfig CR %#q in namespace %#q", certConfig.Name, certConfig.Namespace))
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not create CertConfig CRs")
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

	var certConfigsToCreate []*v1alpha1.CertConfig

	for _, desiredCertConfig := range desiredCertConfigs {
		if !containsCertConfig(currentCertConfigs, desiredCertConfig) {
			certConfigsToCreate = append(certConfigsToCreate, desiredCertConfig)
		}
	}

	return certConfigsToCreate, nil
}
