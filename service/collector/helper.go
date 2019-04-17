package collector

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/service/collector/key"
)

type helperConfig struct {
	G8sClient versioned.Interface
	Logger    micrologger.Logger
}

type helper struct {
	g8sClient versioned.Interface
	logger    micrologger.Logger
}

func newHelper(config helperConfig) (*helper, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	h := &helper{
		g8sClient: config.G8sClient,
		logger:    config.Logger,
	}

	return h, nil
}

func (h *helper) GetTenantClusters() ([]tenantCluster, error) {
	result := []tenantCluster{}

	awsClusters, err := h.getAWSClusters()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	result = append(result, awsClusters...)

	azureClusters, err := h.getAzureClusters()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	result = append(result, azureClusters...)

	kvmClusters, err := h.getKVMClusters()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	result = append(result, kvmClusters...)

	return result, nil
}

func (h *helper) getAWSClusters() ([]tenantCluster, error) {
	result := []tenantCluster{}

	awsClusters, err := h.g8sClient.CoreV1alpha1().AWSClusterConfigs("").List(metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, cr := range awsClusters.Items {
		cluster := tenantCluster{
			apiDomain: key.AWSAPIDomain(cr),
			id:        key.AWSClusterID(cr),
		}

		result = append(result, cluster)
	}

	return result, nil
}

func (h *helper) getAzureClusters() ([]tenantCluster, error) {
	result := []tenantCluster{}

	azureClusters, err := h.g8sClient.CoreV1alpha1().AzureClusterConfigs("").List(metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, cr := range azureClusters.Items {
		cluster := tenantCluster{
			apiDomain: key.AzureAPIDomain(cr),
			id:        key.AzureClusterID(cr),
		}

		result = append(result, cluster)
	}

	return result, nil
}

func (h *helper) getKVMClusters() ([]tenantCluster, error) {
	result := []tenantCluster{}

	azureClusters, err := h.g8sClient.CoreV1alpha1().KVMClusterConfigs("").List(metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, cr := range azureClusters.Items {
		cluster := tenantCluster{
			apiDomain: key.KVMAPIDomain(cr),
			id:        key.KVMClusterID(cr),
		}

		result = append(result, cluster)
	}

	return result, nil
}
