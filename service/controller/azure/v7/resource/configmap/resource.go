package configmap

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"

	"github.com/giantswarm/cluster-operator/pkg/v7/configmap"
)

const (
	// Name is the identifier of the resource.
	Name = "configmapv7"
)

// Config represents the configuration used to create a new chart config resource.
type Config struct {
	ConfigMap configmap.Interface
	Logger    micrologger.Logger

	ProjectName string
}

// Resource implements the chart config resource.
type Resource struct {
	configMap configmap.Interface
	logger    micrologger.Logger

	projectName string
}

// New creates a new configured chart config resource.
func New(config Config) (*Resource, error) {
	if config.ConfigMap == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ConfigMap must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}

	r := &Resource{
		configMap: config.ConfigMap,
		logger:    config.Logger,

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
