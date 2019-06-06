package awsclusterconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v17/key"

	"github.com/giantswarm/microerror"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cluster, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	if !key.IsProviderSpecForAWS(cluster) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("provider extension in cluster cr %q is not for AWS", cluster.Name))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	// TODO: Ensure that managed AWSClusterConfig object is deleted, before
	// removing finalizer from Cluster CR.

	return nil
}
