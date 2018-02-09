package kvmclusterconfig

import (
	"context"

	"github.com/giantswarm/microerror"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	return microerror.New("not implemented")
}
