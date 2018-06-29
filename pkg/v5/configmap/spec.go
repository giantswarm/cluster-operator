package configmap

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

type Interface interface {
	ApplyCreateChange(ctx context.Context, configMapConfig ConfigMapConfig, configMapsToCreate []*corev1.ConfigMap) error
	GetCurrentState(ctx context.Context, configMapConfig ConfigMapConfig) ([]*corev1.ConfigMap, error)
	GetDesiredState(ctx context.Context, configMapValues ConfigMapValues) ([]*corev1.ConfigMap, error)
}

// ConfigMapConfig is used by the configmap resources to provide config to
// calculate the current state.
type ConfigMapConfig struct {
	ClusterID      string
	GuestAPIDomain string
	Namespaces     []string
}

// ConfigMapValues is used by the configmap resources to provide data to the
// configmap service.
type ConfigMapValues struct {
	ClusterID    string
	Organization string
	WorkerCount  int
}
