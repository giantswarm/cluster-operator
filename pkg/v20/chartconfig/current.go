package chartconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
)

func (c *ChartConfig) GetCurrentState(ctx context.Context, clusterConfig ClusterConfig) ([]*v1alpha1.ChartConfig, error) {
	c.logger.LogCtx(ctx, "level", "debug", "message", "looking for chartconfigs in the tenant cluster")

	tenantG8sClient, err := c.newTenantG8sClient(ctx, clusterConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", label.ManagedBy, project.Name()),
	}

	chartConfigList, err := tenantG8sClient.CoreV1alpha1().ChartConfigs(resourceNamespace).List(listOptions)
	if tenant.IsAPINotAvailable(err) {
		c.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is not available yet")
		c.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil

	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	chartConfigs := make([]*v1alpha1.ChartConfig, 0, len(chartConfigList.Items))

	for _, item := range chartConfigList.Items {
		// Make a copy of an Item in order to not refer to loop
		// iterator variable.
		item := item
		chartConfigs = append(chartConfigs, &item)
	}

	c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d chartconfigs in the tenant cluster", len(chartConfigs)))

	return chartConfigs, nil
}
