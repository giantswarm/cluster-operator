package chartconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
)

func (c *ChartConfig) GetCurrentState(ctx context.Context, clusterConfig ClusterConfig) ([]*v1alpha1.ChartConfig, error) {
	return nil, nil
}
