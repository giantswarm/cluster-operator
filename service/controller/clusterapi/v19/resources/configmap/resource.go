package configmap

import (
	"reflect"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
)

const (
	Name = "configmapv19"
)

type Config struct {
	Logger micrologger.Logger
}

type Resource struct {
	logger micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,
	}

	return r, nil
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
