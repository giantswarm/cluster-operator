package kvmclusterconfig

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	return microerror.New("not implemented")
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	return nil, microerror.New("not implemented")
}
