package configmap

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/pkg/v6/configmap"
)

const (
	// Name is the identifier of the resource.
	Name = "configmapv6"
)

// Config represents the configuration used to create a new chart config resource.
type Config struct {
	ConfigMap configmap.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
	Tenant    tenantcluster.Interface

	ProjectName string
}

// Resource implements the chart config resource.
type Resource struct {
	configMap configmap.Interface
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
	tenant    tenantcluster.Interface

	projectName string
}

// New creates a new configured chart config resource.
func New(config Config) (*Resource, error) {
	if config.ConfigMap == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ConfigMap must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Tenant == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Tenant must not be empty", config)
	}

	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}

	r := &Resource{
		configMap: config.ConfigMap,
		k8sClient: config.K8sClient,
		logger:    config.Logger,
		tenant:    config.Tenant,

		projectName: config.ProjectName,
	}

	return r, nil
}

// Name returns name of the Resource.
func (r *Resource) Name() string {
	return Name
}

func toConfigMaps(v interface{}) ([]*corev1.ConfigMap, error) {
	if v == nil {
		return nil, nil
	}

	configMaps, ok := v.([]*corev1.ConfigMap)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []*corev1.ConfigMap{}, v)
	}

	return configMaps, nil
}
