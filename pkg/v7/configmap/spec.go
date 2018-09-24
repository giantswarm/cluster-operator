package configmap

import (
	"context"

	"github.com/giantswarm/operatorkit/controller"
	corev1 "k8s.io/api/core/v1"
)

type Interface interface {
	ApplyCreateChange(ctx context.Context, clusterConfig ClusterConfig, configMapsToCreate []*corev1.ConfigMap) error
	ApplyDeleteChange(ctx context.Context, clusterConfig ClusterConfig, configMapsToDelete []*corev1.ConfigMap) error
	ApplyUpdateChange(ctx context.Context, clusterConfig ClusterConfig, configMapsToUpdate []*corev1.ConfigMap) error
	GetCurrentState(ctx context.Context, configMapConfig ClusterConfig) ([]*corev1.ConfigMap, error)
	GetDesiredState(ctx context.Context, configMapConfig ClusterConfig, configMapValues ConfigMapValues) ([]*corev1.ConfigMap, error)
	NewDeletePatch(ctx context.Context, currentState, desiredState []*corev1.ConfigMap) (*controller.Patch, error)
	NewUpdatePatch(ctx context.Context, currentState, desiredState []*corev1.ConfigMap) (*controller.Patch, error)
}

// ClusterConfig is used by the configmap resources to provide config to
// calculate the current state.
type ClusterConfig struct {
	APIDomain  string
	ClusterID  string
	Namespaces []string
}

// ConfigMapSpec is used to generate the desired state.
type ConfigMapSpec struct {
	App        string
	Labels     map[string]string
	Name       string
	Namespace  string
	ValuesJSON string
}

// ConfigMapValues is used by the configmap resources to provide data to the
// configmap service.
type ConfigMapValues struct {
	CalicoAddress                     string
	CalicoPrefixLength                string
	ClusterID                         string
	ClusterIPRange                    string
	Organization                      string
	IngressControllerMigrationEnabled bool
	IngressControllerUseProxyProtocol bool
	RegistryDomain                    string
	WorkerCount                       int
}

type configMapGenerator func(ctx context.Context, configMapValues ConfigMapValues, projectName string) (*corev1.ConfigMap, error)

// Types below are used for generating values JSON for app configmaps.

type BasicConfigMap struct {
	Image Image `json:"image"`
}

type DefaultConfigMap struct {
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

type ExporterValues struct {
	Namespace string `json:"namespace"`
}

type CertExporter struct {
	Namespace string `json:"namespace"`
}

type NetExporter struct {
	Namespace string `json:"namespace"`
}

type CoreDNS struct {
	Cluster CoreDNSCluster `json:"cluster"`
	Image   Image          `json:"image"`
}

type CoreDNSCluster struct {
	Calico     CoreDNSClusterCalico     `json:"calico"`
	Kubernetes CoreDNSClusterKubernetes `json:"kubernetes"`
}

type CoreDNSClusterCalico struct {
	CIDR string `json:"cidr"`
}

type CoreDNSClusterKubernetes struct {
	API CoreDNSClusterKubernetesAPI `json:"api"`
	DNS CoreDNSClusterKubernetesDNS `json:"dns"`
}

type CoreDNSClusterKubernetesAPI struct {
	ClusterIPRange string `json:"clusterIPRange"`
}

type CoreDNSClusterKubernetesDNS struct {
	IP string `json:"ip"`
}
