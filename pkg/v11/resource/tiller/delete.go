package tiller

import (
	"context"

	"github.com/giantswarm/microerror"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	clusterGuestConfig, err := r.toClusterGuestConfigFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.ensureTillerInstalled(ctx, clusterGuestConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
