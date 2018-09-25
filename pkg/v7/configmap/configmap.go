package configmap

import (
	"context"
	"reflect"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// Config represents the configuration used to create a new configmap service.
type Config struct {
	Logger micrologger.Logger
	Tenant tenantcluster.Interface

	ProjectName    string
	RegistryDomain string
}

// Service provides shared functionality for managing configmaps.
type Service struct {
	logger micrologger.Logger
	tenant tenantcluster.Interface

	projectName    string
	registryDomain string
}

// New creates a new configmap service.
func New(config Config) (*Service, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Tenant == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Tenant must not be empty", config)
	}

	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}
	if config.RegistryDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.RegistryDomain must not be empty", config)
	}

	s := &Service{
		logger: config.Logger,
		tenant: config.Tenant,

		projectName:    config.ProjectName,
		registryDomain: config.RegistryDomain,
	}

	return s, nil
}

func (s *Service) newTenantHelmClient(ctx context.Context, clusterConfig ClusterConfig) (helmclient.Interface, error) {
	tenantK8sClient, err := s.tenant.NewHelmClient(ctx, clusterConfig.ClusterID, clusterConfig.APIDomain)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return tenantK8sClient, nil
}

func (s *Service) newTenantK8sClient(ctx context.Context, clusterConfig ClusterConfig) (kubernetes.Interface, error) {
	tenantK8sClient, err := s.tenant.NewK8sClient(ctx, clusterConfig.ClusterID, clusterConfig.APIDomain)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return tenantK8sClient, nil
}

func containsConfigMap(list []*corev1.ConfigMap, item *corev1.ConfigMap) bool {
	for _, l := range list {
		if reflect.DeepEqual(item, l) {
			return true
		}
	}

	return false
}

func getConfigMapByNameAndNamespace(list []*corev1.ConfigMap, name, namespace string) (*corev1.ConfigMap, error) {
	for _, l := range list {
		if l.Name == name && l.Namespace == namespace {
			return l, nil
		}
	}

	return nil, microerror.Mask(notFoundError)
}

func isConfigMapModified(a, b *corev1.ConfigMap) bool {
	// If the Data section has changed we need to update.
	if !reflect.DeepEqual(a.Data, b.Data) {
		return true
	}
	// If the Labels have changed we also need to update.
	if !reflect.DeepEqual(a.Labels, b.Labels) {
		return true
	}

	return false
}
