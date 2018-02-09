package kvmclusterconfig

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	return microerror.New("not implemented")
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	return nil, microerror.New("not implemented")
}
