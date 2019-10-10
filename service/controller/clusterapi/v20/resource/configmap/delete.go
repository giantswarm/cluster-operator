package configmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v20/controllercontext"
)

// ApplyDeleteChange is executed upon update events in case
// newDeleteChangeForUpdatePatch figured out there are ConfigMap types to be
// deleted.
func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}
	configMaps, err := toConfigMaps(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(configMaps) > 0 {
		for _, configMap := range configMaps {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting configmap %#q in namespace %#q", configMap.Name, configMap.Namespace))

			err := cc.Client.TenantCluster.K8s.CoreV1().ConfigMaps(configMap.Namespace).Delete(configMap.Name, &metav1.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted configmap %#q in namespace %#q", configMap.Name, configMap.Namespace))
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not delete configmaps")
	}

	return nil
}

// NewDeletePatch is a no-op because ConfigMap types in the tenant cluster are
// deleted with the tenant cluster itself upon a delete event.
func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	return nil, nil
}
