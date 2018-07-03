package configmap

import (
	"reflect"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"

	"github.com/giantswarm/cluster-operator/pkg/v5/guestcluster"
)

// Config represents the configuration used to create a new configmap service.
type Config struct {
	Guest  guestcluster.Interface
	Logger micrologger.Logger

	ProjectName string
}

// Service provides shared functionality for managing configmaps.
type Service struct {
	guest  guestcluster.Interface
	logger micrologger.Logger

	projectName string
}

// New creates a new configmap service.
func New(config Config) (*Service, error) {
	if config.Guest == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Guest must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}

	s := &Service{
		guest:       config.Guest,
		logger:      config.Logger,
		projectName: config.ProjectName,
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
