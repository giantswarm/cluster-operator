package configmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	corev1 "k8s.io/api/core/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v20/controllercontext"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}
	configMaps, err := toConfigMaps(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(configMaps) > 0 {
		for _, configMap := range configMaps {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating chartconfig %#q in namespace %#q", configMap.Name, configMap.Namespace))

			_, err := cc.Client.TenantCluster.K8s.CoreV1().ConfigMaps(configMap.Namespace).Update(configMap)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated chartconfig %#q in namespace %#q", configMap.Name, configMap.Namespace))
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not update configmaps")
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
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

	patch := controller.NewPatch()
	patch.SetCreateChange(create)
	patch.SetDeleteChange(delete)
	patch.SetUpdateChange(update)

	return patch, nil
}

// newDeleteChangeForUpdatePatch is specific to the update behaviour because we
// might want to remove certain config maps when a tenant cluster is reconciled.
// So the delete change computed here is gathered for the update patch above.
func (r *Resource) newDeleteChangeForUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) ([]*corev1.ConfigMap, error) {
	currentConfigMaps, err := toConfigMaps(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredConfigMaps, err := toConfigMaps(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var configMapsToDelete []*corev1.ConfigMap

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

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) ([]*corev1.ConfigMap, error) {
	currentConfigMaps, err := toConfigMaps(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredConfigMaps, err := toConfigMaps(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var configMapsToUpdate []*corev1.ConfigMap

	for _, currentConfigMap := range currentConfigMaps {
		desiredConfigMap, err := getConfigMapByNameAndNamespace(desiredConfigMaps, currentConfigMap.Name, currentConfigMap.Namespace)
		if IsNotFound(err) {
			// Ignore here. These are handled by newDeleteChangeForUpdatePatch().
			continue
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		// TODO currently user configmaps are not updated. We should update the
		// metadata. Data keys should not be updated.
		//
		//     https://github.com/giantswarm/giantswarm/issues/4265
		//
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
