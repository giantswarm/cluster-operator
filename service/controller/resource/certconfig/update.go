package certconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/v6/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// applyUpdateChange takes observed custom object and update portion of the
// patch provided by newUpdatePatch or newDeletePatch.
func (r *Resource) applyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	certConfigs, err := toCertConfigs(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(certConfigs) > 0 {
		for _, certConfig := range certConfigs {
			r.logger.Debugf(ctx, "updating CertConfig CR %#q in namespace %#q", certConfig.Name, certConfig.Namespace)

			err = r.ctrlClient.Update(ctx, certConfig, &client.UpdateOptions{Raw: &metav1.UpdateOptions{}})
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.Debugf(ctx, "updated CertConfig CR %#q in namespace %#q", certConfig.Name, certConfig.Namespace)
		}
	} else {
		r.logger.Debugf(ctx, "did not update CertConfig CRs")
	}

	return nil
}

// newUpdatePatch computes appropriate patch based on difference in current
// state and desired state.
func (r *Resource) newUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*patch, error) {
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

	patch := newPatch()
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
