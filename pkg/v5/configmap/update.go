package configmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	corev1 "k8s.io/api/core/v1"
)

func (s *Service) ApplyUpdateChange(ctx context.Context, configMapConfig ConfigMapConfig, configMapsToUpdate []*corev1.ConfigMap) error {
	if len(configMapsToUpdate) > 0 {
		s.logger.LogCtx(ctx, "level", "debug", "message", "updating configmaps")

		guestK8sClient, err := s.guest.NewK8sClient(ctx, configMapConfig.ClusterID, configMapConfig.GuestAPIDomain)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, configMapToUpdate := range configMapsToUpdate {
			_, err := guestK8sClient.CoreV1().ConfigMaps(configMapToUpdate.Namespace).Update(configMapToUpdate)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		s.logger.LogCtx(ctx, "level", "debug", "message", "updated configmaps")
	} else {
		s.logger.LogCtx(ctx, "level", "debug", "message", "no need to update configmaps")
	}

	return nil
}

func (s *Service) NewUpdatePatch(ctx context.Context, currentState, desiredState []*corev1.ConfigMap) (*controller.Patch, error) {
	create, err := s.newCreateChange(ctx, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	update, err := s.newUpdateChange(ctx, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	/* TODO
	delete, err := s.newDeleteChangeForUpdatePatch(ctx, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	*/

	patch := controller.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)
	// patch.SetDeleteChange(delete)

	return patch, nil
}

func (s *Service) newUpdateChange(ctx context.Context, currentConfigMaps, desiredConfigMaps []*corev1.ConfigMap) ([]*corev1.ConfigMap, error) {
	s.logger.LogCtx(ctx, "level", "debug", "message", "finding out which configmaps have to be updated")

	configMapsToUpdate := make([]*corev1.ConfigMap, 0)

	for _, currentConfigMap := range currentConfigMaps {
		desiredConfigMap, err := getConfigMapByNameAndNamespace(desiredConfigMaps, currentConfigMap.Name, currentConfigMap.Namespace)
		if IsNotFound(err) {
			// Ignore here. These are handled by newDeleteChangeForUpdatePatch().
			continue
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		if isConfigMapModified(desiredConfigMap, currentConfigMap) {
			// Make a copy and set the resource version so the CR can be updated.
			configMapToUpdate := desiredConfigMap.DeepCopy()
			configMapToUpdate.ObjectMeta.ResourceVersion = currentConfigMap.ObjectMeta.ResourceVersion

			configMapsToUpdate = append(configMapsToUpdate, configMapToUpdate)

			s.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found configmap '%s' that has to be updated", desiredConfigMap.GetName()))
		}
	}

	s.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d configmaps which have to be updated", len(configMapsToUpdate)))

	return configMapsToUpdate, nil
}
