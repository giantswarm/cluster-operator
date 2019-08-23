package configmap

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, clusterConfig ClusterConfig, configMapsToCreate []*corev1.ConfigMap) error {
	if len(configMapsToCreate) > 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating configmaps")

		for _, configMapToCreate := range configMapsToCreate {
			_, err := cc.Client.TenantCluster.K8s.CoreV1().ConfigMaps(configMapToCreate.Namespace).Create(configMapToCreate)
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "created configmaps")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no need to create configmaps")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, currentConfigMaps, desiredConfigMaps []*corev1.ConfigMap) ([]*corev1.ConfigMap, error) {

	configMapsToCreate := make([]*corev1.ConfigMap, 0)

	for _, desiredConfigMap := range desiredConfigMaps {
		if !containsConfigMap(currentConfigMaps, desiredConfigMap) {
			configMapsToCreate = append(configMapsToCreate, desiredConfigMap)
		}
	}

	return configMapsToCreate, nil
}
