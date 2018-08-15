package configmap

import (
	"context"

	"github.com/giantswarm/operatorkit/controller"
	corev1 "k8s.io/api/core/v1"
)

type Interface interface {
	ApplyCreateChange(ctx context.Context, configMapConfig ConfigMapConfig, configMapsToCreate []*corev1.ConfigMap) error
	ApplyDeleteChange(ctx context.Context, configMapConfig ConfigMapConfig, configMapsToDelete []*corev1.ConfigMap) error
	ApplyUpdateChange(ctx context.Context, configMapConfig ConfigMapConfig, configMapsToUpdate []*corev1.ConfigMap) error
	GetCurrentState(ctx context.Context, configMapConfig ConfigMapConfig) ([]*corev1.ConfigMap, error)
	GetDesiredState(ctx context.Context, configMapValues ConfigMapValues) ([]*corev1.ConfigMap, error)
	NewDeletePatch(ctx context.Context, currentState, desiredState []*corev1.ConfigMap) (*controller.Patch, error)
	NewUpdatePatch(ctx context.Context, currentState, desiredState []*corev1.ConfigMap) (*controller.Patch, error)
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
	ClusterID                         string
	Organization                      string
	IngressControllerMigrationEnabled bool
	WorkerCount                       int
}
