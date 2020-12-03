package appresource

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	appCRs, err := toAppCRs(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, appCR := range appCRs {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating App CR %#q in namespace %#q", appCR.Name, appCR.Namespace))

		// Get app CR again to ensure the resource version is correct.
		currentCR, err := r.g8sClient.ApplicationV1alpha1().Apps(appCR.Namespace).Get(ctx, appCR.Name, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		appCR.ResourceVersion = currentCR.ResourceVersion

		_, err = r.g8sClient.ApplicationV1alpha1().Apps(appCR.Namespace).Update(appCR)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated App CR %#q in namespace %#q", appCR.Name, appCR.Namespace))
	}

	return nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentAppCRs, err := toAppCRs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredAppCRs, err := toAppCRs(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var appCRsToUpdate []*v1alpha1.App
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("computing App CRs to update"))

		for _, c := range currentAppCRs {
			for _, d := range desiredAppCRs {
				m := newAppCRToUpdate(c, d, r.allowedAnnotations)
				if m != nil {
					appCRsToUpdate = append(appCRsToUpdate, m)
				}
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("computed %d App CRs to update", len(appCRsToUpdate)))
	}

	return appCRsToUpdate, nil
}
