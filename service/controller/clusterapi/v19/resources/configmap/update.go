package configmap

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	corev1 "k8s.io/api/core/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, clusterConfig ClusterConfig, configMapsToUpdate []*corev1.ConfigMap) error {
	if len(configMapsToUpdate) > 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating configmaps")

		for _, configMapToUpdate := range configMapsToUpdate {
			_, err := cc.Client.TenantCluster.K8s.CoreV1().ConfigMaps(configMapToUpdate.Namespace).Update(configMapToUpdate)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated configmaps")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no need to update configmaps")
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, currentState, desiredState []*corev1.ConfigMap) (*controller.Patch, error) {
	create, err := r.newCreateChange(ctx, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	delete, err := r.newDeleteChangeForUpdatePatch(ctx, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	update, err := r.newUpdateChange(ctx, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetCreateChange(create)
	patch.SetDeleteChange(delete)
	patch.SetUpdateChange(update)

	return patch, nil
}

func (r *Resource) newDeleteChangeForUpdatePatch(ctx context.Context, currentConfigMaps, desiredConfigMaps []*corev1.ConfigMap) ([]*corev1.ConfigMap, error) {
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

	return configMapsToDelete, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, currentConfigMaps, desiredConfigMaps []*corev1.ConfigMap) ([]*corev1.ConfigMap, error) {
	configMapsToUpdate := make([]*corev1.ConfigMap, 0)

	for _, currentConfigMap := range currentConfigMaps {
		desiredConfigMap, err := getConfigMapByNameAndNamespace(desiredConfigMaps, currentConfigMap.Name, currentConfigMap.Namespace)
		if IsNotFound(err) {
			// Ignore here. These are handled by newDeleteChangeForUpdatePatch().
			continue
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		// Currently user configmaps are not updated. We should update the
		// metadata. Data keys should not be updated.
		// TODO https://github.com/giantswarm/giantswarm/issues/4265
		configMapType := currentConfigMap.Labels[label.ConfigMapType]
		if configMapType != label.ConfigMapTypeUser {
			if isConfigMapModified(desiredConfigMap, currentConfigMap) {
				// Make a copy and set the resource version so the CR can be updated.
				configMapToUpdate := desiredConfigMap.DeepCopy()
				configMapToUpdate.ObjectMeta.ResourceVersion = currentConfigMap.ObjectMeta.ResourceVersion

				configMapsToUpdate = append(configMapsToUpdate, configMapToUpdate)

			}
		}
	}

	return configMapsToUpdate, nil
}
