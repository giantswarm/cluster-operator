package configmap

import (
	"context"

	"github.com/giantswarm/operatorkit/controller"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	patch := controller.NewPatch()

	return patch, nil
}
