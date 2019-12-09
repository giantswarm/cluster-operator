package basedomain

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	// TODO fetch tenant cluster base domain in a generic way from AWSCluster CRs
	// as referenced by Cluster CRs.
	cc.Status.Endpoint.Base = ""

	return nil
}
