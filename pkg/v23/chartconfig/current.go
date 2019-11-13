package chartconfig

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/pkg/v23/controllercontext"
)

func (c *ChartConfig) GetCurrentState(ctx context.Context, clusterConfig ClusterConfig) ([]*v1alpha1.ChartConfig, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	{
		if cc.Client.TenantCluster.G8s == nil {
			c.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster clients not available")
			c.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil, nil
		}
	}

	c.logger.LogCtx(ctx, "level", "debug", "message", "looking for chartconfigs in the tenant cluster")

	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", label.ManagedBy, project.Name()),
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	chartConfigList, err := cc.Client.TenantCluster.G8s.CoreV1alpha1().ChartConfigs(resourceNamespace).List(listOptions)
	if tenant.IsAPINotAvailable(err) {
		c.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is not available yet")
		c.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		c.logger.LogCtx(ctx, "level", "debug", "message", "timeout getting chartconfig CRs")
		c.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil
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
