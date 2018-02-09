package kvmclusterconfig

import (
	"context"

	"github.com/giantswarm/microerror"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	return nil, microerror.New("not implemented")
}
