package configmap

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

type Interface interface {
	GetDesiredState(ctx context.Context, configMapValues ConfigMapValues) ([]*corev1.ConfigMap, error)
}

// ConfigMapValues is used by the configmap resources to provide data to the
// configmap service.
type ConfigMapValues struct {
	ClusterID    string
	Organization string
	WorkerCount  int
}
