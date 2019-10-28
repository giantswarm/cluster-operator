package configmap

import (
	"reflect"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
)

const (
	Name = "configmapv22"
)

type Config struct {
	Logger micrologger.Logger

	// CalicoAddress may be empty on certain installations.
	CalicoAddress string
	// CalicoPrefixLength may be empty on certain installations.
	CalicoPrefixLength string
	ClusterIPRange     string
	DNSIP              string
	Provider           string
	RegistryDomain     string
}

type Resource struct {
	logger micrologger.Logger

	calicoAddress      string
	calicoPrefixLength string
	clusterIPRange     string
	dnsIP              string
	provider           string
	registryDomain     string
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ClusterIPRange == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterIPRange must not be empty", config)
	}
	if config.DNSIP == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.DNSIP must not be empty", config)
	}
	if config.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}
	if config.RegistryDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.RegistryDomain must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,

		calicoAddress:      config.CalicoAddress,
		calicoPrefixLength: config.CalicoPrefixLength,
		clusterIPRange:     config.ClusterIPRange,
		dnsIP:              config.DNSIP,
		provider:           config.Provider,
		registryDomain:     config.RegistryDomain,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

// containsConfigMap checks if item is present within list
// by comparing ObjectMeta Name and Namespace property between item and list objects.
func containsConfigMap(list []*corev1.ConfigMap, item *corev1.ConfigMap) bool {
	for _, l := range list {
		if item.Name == l.Name && item.Namespace == l.Namespace {
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

func toConfigMaps(v interface{}) ([]*corev1.ConfigMap, error) {
	if v == nil {
		return nil, nil
	}

	t, ok := v.([]*corev1.ConfigMap)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", t, v)
	}

	return t, nil
}
