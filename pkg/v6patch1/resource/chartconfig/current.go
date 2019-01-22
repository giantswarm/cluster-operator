package chartconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/errors/guest"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"github.com/giantswarm/tenantcluster"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetCurrentState returns the ChartConfig resources present in the guest
// cluster.
func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {

	r.logger.LogCtx(ctx, "level", "debug", "message", "looking for chartconfigs in the guest cluster")

	tenantG8sClient, err := r.getTenantG8sClient(ctx, obj)
	if tenantcluster.IsTimeout(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the cluster-operator api cert in the Kubernetes API")

		// We can't continue without the cert. We will retry during the next
		// execution.
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil, nil

	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	chartConfigList, err := tenantG8sClient.CoreV1alpha1().ChartConfigs(resourceNamespace).List(metav1.ListOptions{})
	if apierrors.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the chartconfig CRD in the guest cluster")

		// ChartConfig CRD is created by chart-operator which may not be
		// running yet in the guest cluster. We will retry during the next
		// execution.
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil, nil
	} else if guest.IsAPINotAvailable(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "guest cluster is not available")

		// We can't continue without a successful K8s connection. Cluster
		// may not be up yet. We will retry during the next execution.
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

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

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d chartconfigs in the guest cluster", len(chartConfigs)))

	return chartConfigs, nil
}
