package configmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/resource/crud"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Service) ApplyDeleteChange(ctx context.Context, clusterConfig ClusterConfig, configMapsToDelete []*corev1.ConfigMap) error {
	if len(configMapsToDelete) > 0 {
		s.logger.LogCtx(ctx, "level", "debug", "message", "deleting configmaps")

		tenantK8sClient, err := s.newTenantK8sClient(ctx, clusterConfig)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, configMap := range configMapsToDelete {
			err := tenantK8sClient.CoreV1().ConfigMaps(configMap.Namespace).Delete(configMap.Name, &metav1.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		s.logger.LogCtx(ctx, "level", "debug", "message", "deleted configmaps")
	} else {
		s.logger.LogCtx(ctx, "level", "debug", "message", "no need to delete configmaps")
	}

	return nil
}

// NewDeletePatch is a no-op because configmaps in the tenant cluster are
// deleted with the tenant cluster resources.
func (s *Service) NewDeletePatch(ctx context.Context, currentState, desiredState []*corev1.ConfigMap) (*crud.Patch, error) {
	return nil, nil
}

func (s *Service) newDeleteChangeForUpdatePatch(ctx context.Context, currentConfigMaps, desiredConfigMaps []*corev1.ConfigMap) ([]*corev1.ConfigMap, error) {
	s.logger.LogCtx(ctx, "level", "debug", "message", "finding out which configmaps have to be deleted")

	configMapsToDelete := make([]*corev1.ConfigMap, 0)

	for _, currentConfigMap := range currentConfigMaps {
		_, err := getConfigMapByNameAndNamespace(desiredConfigMaps, currentConfigMap.Name, currentConfigMap.Namespace)
		// Existing ConfigMap is not desired anymore so it should be deleted.
		if IsNotFound(err) {
			configMapsToDelete = append(configMapsToDelete, currentConfigMap)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d configmaps that have to be deleted", len(configMapsToDelete)))

	return configMapsToDelete, nil
}
