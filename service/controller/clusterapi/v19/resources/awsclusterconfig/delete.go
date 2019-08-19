package awsclusterconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting AWSClusterConfig %#q", key.AWSClusterConfigName(cr)))

	if !key.IsProviderSpecForAWS(cr) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("provider extension in cluster %#q is not for AWS", cr.Name))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	err = r.g8sClient.CoreV1alpha1().AWSClusterConfigs(cr.Namespace).Delete(key.AWSClusterConfigName(cr), &metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		// fall through
	} else if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted AWSClusterConfig %#q", key.AWSClusterConfigName(cr)))

	return nil
}
