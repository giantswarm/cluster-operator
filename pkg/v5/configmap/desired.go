package configmap

import (
	"context"
	"encoding/json"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
)

func (s *Service) GetDesiredState(ctx context.Context, configMapValues ConfigMapValues) ([]*corev1.ConfigMap, error) {
	desiredConfigMaps := make([]*corev1.ConfigMap, 0)

	{
		configMap, err := s.newIngressControllerConfigMap(ctx, configMapValues, s.projectName)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		desiredConfigMaps = append(desiredConfigMaps, configMap)
	}

	return desiredConfigMaps, nil
}

type IngressController struct {
	Controller IngressControllerController `json:"controller"`
}

type IngressControllerController struct {
	ReplicaCount int `json:"replicaCount"`
}

func (s *Service) newIngressControllerConfigMap(ctx context.Context, configMapValues ConfigMapValues, projectName string) (*corev1.ConfigMap, error) {
	configMapName := "nginx-ingress-controller-values"
	appName := "nginx-ingress-controller"
	labels := newConfigMapLabels(configMapValues, appName, projectName)

	values := IngressController{
		Controller: IngressControllerController{
			ReplicaCount: configMapValues.WorkerCount,
		},
	}
	json, err := json.Marshal(values)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	data := map[string]string{
		"values.json": string(json),
	}

	newConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: metav1.NamespaceSystem,
			Labels:    labels,
		},
		Data: data,
	}

	return newConfigMap, nil
}

func newConfigMapLabels(configMapValues ConfigMapValues, appName, projectName string) map[string]string {
	return map[string]string{
		label.App:          appName,
		label.Cluster:      configMapValues.ClusterID,
		label.ManagedBy:    projectName,
		label.Organization: configMapValues.Organization,
		label.ServiceType:  label.ServiceTypeManaged,
	}
}
