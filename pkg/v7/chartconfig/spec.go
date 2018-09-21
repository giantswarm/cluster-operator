package chartconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"

	"github.com/giantswarm/cluster-operator/pkg/v7/key"
)

type Interface interface {
	GetCurrentState(ctx context.Context, clusterConfig ClusterConfig) ([]*v1alpha1.ChartConfig, error)
	GetDesiredState(ctx context.Context, clusterConfig ClusterConfig, providerChartSpecs []key.ChartSpec) ([]*v1alpha1.ChartConfig, error)
}

// ClusterConfig is used by the chartconfig service to provide config to
// connect to the tenant cluster.
type ClusterConfig struct {
	APIDomain    string
	ClusterID    string
	Organization string
}
