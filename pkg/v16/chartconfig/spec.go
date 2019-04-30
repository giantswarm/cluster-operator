package chartconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/operatorkit/controller"

	"github.com/giantswarm/cluster-operator/pkg/v15/key"
)

type Interface interface {
	ApplyCreateChange(ctx context.Context, clusterConfig ClusterConfig, chartConfigsToCreate []*v1alpha1.ChartConfig) error
	ApplyDeleteChange(ctx context.Context, clusterConfig ClusterConfig, chartConfigsToDelete []*v1alpha1.ChartConfig) error
	ApplyUpdateChange(ctx context.Context, clusterConfig ClusterConfig, chartConfigsToUpdate []*v1alpha1.ChartConfig) error
	GetCurrentState(ctx context.Context, clusterConfig ClusterConfig) ([]*v1alpha1.ChartConfig, error)
	GetDesiredState(ctx context.Context, clusterConfig ClusterConfig, providerChartSpecs []key.ChartSpec) ([]*v1alpha1.ChartConfig, error)
	NewUpdatePatch(ctx context.Context, currentState, desiredState []*v1alpha1.ChartConfig) (*controller.Patch, error)
	NewDeletePatch(ctx context.Context, currentState, desiredState []*v1alpha1.ChartConfig) (*controller.Patch, error)
}

// ClusterConfig is used by the chartconfig service to provide config to
// connect to the tenant cluster.
type ClusterConfig struct {
	APIDomain    string
	ClusterID    string
	Organization string
}
