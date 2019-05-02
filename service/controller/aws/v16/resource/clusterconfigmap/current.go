package clusterconfigmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/v16/key"
	awskey "github.com/giantswarm/cluster-operator/service/controller/aws/v16/key"
)

func (r *StateGetter) GetCurrentState(ctx context.Context, obj interface{}) ([]*v1.ConfigMap, error) {
	cr, err := awskey.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterConfig := awskey.ClusterGuestConfig(cr)
	name := key.ClusterConfigMapName(clusterConfig)

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding cluster configMap %#q in namespace %#q", name, clusterConfig.ID))

	cm, err := r.k8sClient.CoreV1().ConfigMaps(clusterConfig.ID).Get(name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find cluster configMap %#q in namespace %#q", name, clusterConfig.ID))
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found cluster configMap %#q in namespace %#q", name, clusterConfig.ID))

	return []*v1.ConfigMap{cm}, nil
}
