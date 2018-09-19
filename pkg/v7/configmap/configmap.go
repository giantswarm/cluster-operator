package configmap

import (
	"reflect"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster"
	corev1 "k8s.io/api/core/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
)

// Config represents the configuration used to create a new configmap service.
type Config struct {
	Logger micrologger.Logger
	Tenant tenantcluster.Interface

	CalicoAddress      string
	CalicoPrefixLength string
	ClusterIPRange     string
	ProjectName        string
	Provider           string
	RegistryDomain     string
}

// Service provides shared functionality for managing configmaps.
type Service struct {
	logger micrologger.Logger
	tenant tenantcluster.Interface

	calicoAddress      string
	calicoPrefixLength string
	clusterIPRange     string
	projectName        string
	provider           string
	registryDomain     string
}

// New creates a new configmap service.
func New(config Config) (*Service, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Tenant == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Tenant must not be empty", config)
	}

	// Azure manages Calico CIDR blocks differently, and the installations
	// settings can be empty.
	if config.Provider != label.ProviderAzure {
		if config.CalicoAddress == "" {
			return nil, microerror.Maskf(invalidConfigError, "%T.CalicoAddress must not be empty", config)
		}
		if config.CalicoPrefixLength == "" {
			return nil, microerror.Maskf(invalidConfigError, "%T.CalicoPrefixLength must not be empty", config)
		}
	}

	if config.ClusterIPRange == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterIPRange must not be empty", config)
	}

	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}

	if config.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	if config.RegistryDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.RegistryDomain must not be empty", config)
	}

	s := &Service{
		logger: config.Logger,
		tenant: config.Tenant,

		calicoAddress:      config.CalicoAddress,
		calicoPrefixLength: config.CalicoPrefixLength,
		clusterIPRange:     config.ClusterIPRange,
		projectName:        config.ProjectName,
		provider:           config.Provider,
		registryDomain:     config.RegistryDomain,
	}

	return s, nil
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
