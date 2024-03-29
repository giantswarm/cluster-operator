package certconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/v6/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
			r.logger.Debugf(ctx, "creating CertConfig CR %#q in namespace %#q", certConfig.Name, certConfig.Namespace)

			err = r.ctrlClient.Create(ctx, certConfig, &client.CreateOptions{Raw: &metav1.CreateOptions{}})
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}

			r.logger.Debugf(ctx, "created CertConfig CR %#q in namespace %#q", certConfig.Name, certConfig.Namespace)
		}
	} else {
		r.logger.Debugf(ctx, "did not create CertConfig CRs")
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
