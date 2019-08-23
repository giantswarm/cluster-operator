package chartconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	{
		if cc.Client.TenantCluster.G8s == nil {
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster clients not available")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil, nil
		}
	}

	var chartConfigs []*v1alpha1.ChartConfig
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding chart configs in tenant cluster %#q", key.ClusterID(&cr)))

		o := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", label.ManagedBy, project.Name()),
		}

		list, err := cc.Client.TenantCluster.G8s.CoreV1alpha1().ChartConfigs("giantswarm").List(o)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		for _, item := range list.Items {
			// Make a copy of an Item in order to not refer to loop iterator
			// variable. This is because we want to track a list of pointers.
			item := item
			chartConfigs = append(chartConfigs, &item)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d chart configs in tenant cluster %#q", len(chartConfigs), key.ClusterID(&cr)))
	}

	return chartConfigs, nil
}
