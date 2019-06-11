package awsclusterconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v17/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cluster, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting AWSClusterConfig %q", key.AWSClusterConfigName(cluster)))

	if !key.IsProviderSpecForAWS(cluster) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("provider extension in cluster cr %q is not for AWS", cluster.Name))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	err = r.g8sClient.CoreV1alpha1().AWSClusterConfigs(cluster.Namespace).Delete(key.AWSClusterConfigName(cluster), &metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		// fall through
	} else if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted AWSClusterConfig %q", key.AWSClusterConfigName(cluster)))

	return nil
}
