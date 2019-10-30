package configmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/cluster-operator/pkg/label"
)

func (s *Service) ApplyCreateChange(ctx context.Context, clusterConfig ClusterConfig, configMapsToCreate []*corev1.ConfigMap) error {
	if len(configMapsToCreate) > 0 {
		s.logger.LogCtx(ctx, "level", "debug", "message", "creating configmaps")

		tenantK8sClient, err := s.newTenantK8sClient(ctx, clusterConfig)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, configMapToCreate := range configMapsToCreate {
			_, err := tenantK8sClient.CoreV1().ConfigMaps(configMapToCreate.Namespace).Create(configMapToCreate)
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		s.logger.LogCtx(ctx, "level", "debug", "message", "created configmaps")
	} else {
		s.logger.LogCtx(ctx, "level", "debug", "message", "no need to create configmaps")
	}

	return nil
}

func (s *Service) newCreateChange(ctx context.Context, currentConfigMaps, desiredConfigMaps []*corev1.ConfigMap) ([]*corev1.ConfigMap, error) {
	s.logger.LogCtx(ctx, "level", "debug", "message", "finding out which chartconfigs have to be created")

	configMapsToCreate := make([]*corev1.ConfigMap, 0)

	for _, desiredConfigMap := range desiredConfigMaps {
		appName := desiredConfigMap.Labels[label.App]
		chartSpec := s.getChartSpecByAppName(appName)
		if chartSpec.HasAppCR {
			continue
		}

		if !containsConfigMap(currentConfigMaps, desiredConfigMap) {
			configMapsToCreate = append(configMapsToCreate, desiredConfigMap)
		}
	}

	s.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d configmaps that have to be created", len(configMapsToCreate)))

	return configMapsToCreate, nil
}
