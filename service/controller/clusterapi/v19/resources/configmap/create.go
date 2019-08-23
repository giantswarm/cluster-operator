package configmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/controllercontext"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}
	configMaps, err := toConfigMaps(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(configMaps) > 0 {
		for _, configMap := range configMaps {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating configmap %#q in namespace %#q", configMap.Name, configMap.Namespace))

			_, err := cc.Client.TenantCluster.K8s.CoreV1().ConfigMaps(configMap.Namespace).Create(configMap)
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created configmap %#q in namespace %#q", configMap.Name, configMap.Namespace))
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not create configmaps")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) ([]*corev1.ConfigMap, error) {
	currentConfigMaps, err := toConfigMaps(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredConfigMaps, err := toConfigMaps(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var configMapsToCreate []*corev1.ConfigMap

	for _, desiredConfigMap := range desiredConfigMaps {
		if !containsConfigMap(currentConfigMaps, desiredConfigMap) {
			configMapsToCreate = append(configMapsToCreate, desiredConfigMap)
		}
	}

	return configMapsToCreate, nil
}
