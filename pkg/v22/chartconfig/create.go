package chartconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/cluster-operator/pkg/v21/controllercontext"
)

func (c *ChartConfig) ApplyCreateChange(ctx context.Context, clusterConfig ClusterConfig, chartConfigsToCreate []*v1alpha1.ChartConfig) error {
	if len(chartConfigsToCreate) > 0 {
		cc, err := controllercontext.FromContext(ctx)
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", "creating chartconfigs")

		for _, chartConfigToCreate := range chartConfigsToCreate {
			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating chartconfig %#q", chartConfigToCreate.Name))

			_, err := cc.Client.TenantCluster.G8s.CoreV1alpha1().ChartConfigs(resourceNamespace).Create(chartConfigToCreate)
			if apierrors.IsAlreadyExists(err) {
				c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not create chartconfig %#q", chartConfigToCreate.Name))
				c.logger.LogCtx(ctx, "level", "debug", "message", "chartconfig already exists")
			} else if err != nil {
				return microerror.Mask(err)
			}

			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created chartconfig %#q", chartConfigToCreate.Name))
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", "created chartconfigs")
	} else {
		c.logger.LogCtx(ctx, "level", "debug", "message", "no need to create chartconfigs")
	}

	return nil
}

func (c *ChartConfig) newCreateChange(ctx context.Context, currentChartConfigs, desiredChartConfigs []*v1alpha1.ChartConfig) ([]*v1alpha1.ChartConfig, error) {
	c.logger.LogCtx(ctx, "level", "debug", "message", "finding out which chartconfigs have to be created")

	chartConfigsToCreate := make([]*v1alpha1.ChartConfig, 0)

	for _, desiredChartConfig := range desiredChartConfigs {
		chartSpec := c.getChartSpecByName(desiredChartConfig.Name)
		if chartSpec.HasAppCR {
			continue
		}

		if !containsChartConfig(currentChartConfigs, desiredChartConfig) {
			chartConfigsToCreate = append(chartConfigsToCreate, desiredChartConfig)
		}
	}

	c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d chartconfigs that have to be created", len(chartConfigsToCreate)))

	return chartConfigsToCreate, nil
}
