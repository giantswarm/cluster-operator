package chartconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/operatorkit/controller"
)

func (c *ChartConfig) ApplyDeleteChange(ctx context.Context, clusterConfig ClusterConfig, chartConfigsToDelete []*v1alpha1.ChartConfig) error {
	return nil
}

func (c *ChartConfig) NewDeletePatch(ctx context.Context, currentState, desiredState []*v1alpha1.ChartConfig) (*controller.Patch, error) {
	return nil, nil
}
