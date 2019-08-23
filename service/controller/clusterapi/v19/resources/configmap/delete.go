package configmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, clusterConfig ClusterConfig, configMapsToDelete []*corev1.ConfigMap) error {
	if len(configMapsToDelete) > 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting configmaps")

		for _, configMap := range configMapsToDelete {
			err := cc.Client.TenantCluster.K8s.CoreV1().ConfigMaps(configMap.Namespace).Delete(configMap.Name, &metav1.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted configmap %#q in namespace %#q", chartConfig.Name, chartConfig.Namespace))
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not delete configmaps")
	}

	return nil
}

// NewDeletePatch is a no-op because configmaps in the tenant cluster are
// deleted with the tenant cluster resources.
func (r *Resource) NewDeletePatch(ctx context.Context, currentState, desiredState []*corev1.ConfigMap) (*controller.Patch, error) {
	return nil, nil
}
