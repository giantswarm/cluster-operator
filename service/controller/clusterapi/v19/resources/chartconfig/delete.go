package chartconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/controllercontext"
)

// ApplyDeleteChange is executed upon update events in case
// newDeleteChangeForUpdatePatch figured out there are ChartConfig CRs to be
// deleted.
func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}
	chartConfigs, err := toChartConfigs(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(chartConfigs) > 0 {
		for _, chartConfig := range chartConfigs {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting chartconfig %#q in namespace %#q", chartConfig.Name, chartConfig.Namespace))

			err := cc.Client.TenantCluster.G8s.CoreV1alpha1().ChartConfigs(chartConfig.Namespace).Delete(chartConfig.Name, &metav1.DeleteOptions{})
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted chartconfig %#q in namespace %#q", chartConfig.Name, chartConfig.Namespace))
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not delete chartconfigs")
	}

	return nil
}

// NewDeletePatch is a no-op because ChartConfig CRs in the tenant cluster are
// deleted with the tenant cluster itself upon a delete event.
func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	return nil, nil
}
