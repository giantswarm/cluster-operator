package configmap

import (
	"context"
	"encoding/json"
	"math"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
)

type configMapGenerator func(ctx context.Context, configMapValues ConfigMapValues, projectName string) (*corev1.ConfigMap, error)

type BasicConfigMap struct {
	Image Image `json:"image"`
}

type IngressController struct {
	Controller IngressControllerController `json:"controller"`
	Global     IngressControllerGlobal     `json:"global"`
	Image      Image                       `json:"image"`
}

type IngressControllerController struct {
	Replicas int                                `json:"replicas"`
	Service  IngressControllerControllerService `json:"service"`
}

type IngressControllerControllerService struct {
	Enabled bool `json:"enabled"`
}

type IngressControllerGlobal struct {
	Controller IngressControllerGlobalController `json:"controller"`
	Migration  IngressControllerGlobalMigration  `json:"migration"`
}

type IngressControllerGlobalController struct {
	TempReplicas     int  `json:"tempReplicas"`
	UseProxyProtocol bool `json:"useProxyProtocol"`
}

type IngressControllerGlobalMigration struct {
	Enabled bool `json:"enabled"`
}

type Image struct {
	Registry string `json:"registry"`
}

type CertExporter struct {
	Namespace string `json:"namespace"`
}

type NetExporter struct {
	Namespace string `json:"namespace"`
}

func (s *Service) GetDesiredState(ctx context.Context, configMapConfig ConfigMapConfig, configMapValues ConfigMapValues) ([]*corev1.ConfigMap, error) {
	desiredConfigMaps := make([]*corev1.ConfigMap, 0)

	configMap, err := s.newIngressControllerConfigMap(ctx, configMapConfig, configMapValues, s.projectName)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredConfigMaps = append(desiredConfigMaps, configMap)
	generators := []configMapGenerator{
		s.newCertExporterConfigMap,
		s.newKubeStateMetricsConfigMap,
		s.newNetExporterConfigMap,
		s.newNodeExporterConfigMap,
	}

	for _, g := range generators {
		configMap, err := g(ctx, configMapValues, s.projectName)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		desiredConfigMaps = append(desiredConfigMaps, configMap)
	}

	return desiredConfigMaps, nil
}

func (s *Service) newCertExporterConfigMap(ctx context.Context, configMapValues ConfigMapValues, projectName string) (*corev1.ConfigMap, error) {
	configMapName := "cert-exporter-values"
	appName := "cert-exporter"
	labels := newConfigMapLabels(configMapValues, appName, projectName)

	values := CertExporter{
		Namespace: metav1.NamespaceSystem,
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

func (s *Service) newIngressControllerConfigMap(ctx context.Context, configMapConfig ConfigMapConfig, configMapValues ConfigMapValues, projectName string) (*corev1.ConfigMap, error) {
	configMapName := "nginx-ingress-controller-values"
	appName := "nginx-ingress-controller"
	labels := newConfigMapLabels(configMapValues, appName, projectName)

	// controllerServiceEnabled needs to be set separately for the chart
	// migration logic but is the reverse of migration enabled.
	controllerServiceEnabled := !configMapValues.IngressControllerMigrationEnabled

	migrationEnabled := configMapValues.IngressControllerMigrationEnabled
	if migrationEnabled {
		releaseExists, err := s.checkHelmReleaseExists(ctx, appName, configMapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		// Release exists so don't repeat the migration process.
		if releaseExists {
			migrationEnabled = false
		}
	}

	// tempReplicas is set to 50% of the worker count to ensure all pods can be
	// scheduled.
	tempReplicas, err := setIngressControllerTempReplicas(configMapValues.WorkerCount)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	values := IngressController{
		Controller: IngressControllerController{
			Replicas: configMapValues.WorkerCount,
			Service: IngressControllerControllerService{
				Enabled: controllerServiceEnabled,
			},
		},
		Global: IngressControllerGlobal{
			Controller: IngressControllerGlobalController{
				TempReplicas:     tempReplicas,
				UseProxyProtocol: configMapValues.IngressControllerUseProxyProtocol,
			},
			Migration: IngressControllerGlobalMigration{
				Enabled: migrationEnabled,
			},
		},
		Image: Image{
			Registry: s.registryDomain,
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

func (s *Service) newKubeStateMetricsConfigMap(ctx context.Context, configMapValues ConfigMapValues, projectName string) (*corev1.ConfigMap, error) {
	return s.newBasicConfigMap(ctx, configMapValues, projectName, "kube-state-metrics")
}

func (s *Service) newNetExporterConfigMap(ctx context.Context, configMapValues ConfigMapValues, projectName string) (*corev1.ConfigMap, error) {
	configMapName := "net-exporter-values"
	appName := "net-exporter"
	labels := newConfigMapLabels(configMapValues, appName, projectName)

	values := NetExporter{
		Namespace: metav1.NamespaceSystem,
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

func (s *Service) newNodeExporterConfigMap(ctx context.Context, configMapValues ConfigMapValues, projectName string) (*corev1.ConfigMap, error) {
	return s.newBasicConfigMap(ctx, configMapValues, projectName, "node-exporter")
}

func (s *Service) newBasicConfigMap(ctx context.Context, configMapValues ConfigMapValues, projectName string, appName string) (*corev1.ConfigMap, error) {
	configMapName := appName + "-values"
	labels := newConfigMapLabels(configMapValues, appName, projectName)

	values := BasicConfigMap{
		Image: Image{
			Registry: s.registryDomain,
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

func (s *Service) checkHelmReleaseExists(ctx context.Context, releaseName string, configMapConfig ConfigMapConfig) (bool, error) {
	guestHelmClient, err := s.guest.NewHelmClient(ctx, configMapConfig.ClusterID, configMapConfig.GuestAPIDomain)
	if err != nil {
		return false, microerror.Mask(err)
	}

	_, err = guestHelmClient.GetReleaseContent(ctx, releaseName)
	if helmclient.IsReleaseNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
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

// setIngressControllerTempReplicas sets the temp replicas to 50% of the worker
// count to ensure all pods can be scheduled.
func setIngressControllerTempReplicas(workerCount int) (int, error) {
	if workerCount == 0 {
		return 0, microerror.Maskf(invalidExecutionError, "worker count must not be 0")
	}

	tempReplicas := float64(workerCount) * float64(0.5)

	return int(math.Round(tempReplicas)), nil
}
