package basedomain

import (
	"context"
	"fmt"

	"github.com/giantswarm/api/pkg/key"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
)

const (
	Name = "basedomain"
)

type Config struct {
	Logger        micrologger.Logger
	ToClusterFunc func(ctx context.Context, obj interface{}) (apiv1alpha2.Cluster, error)
}

type Resource struct {
	logger        micrologger.Logger
	toClusterFunc func(ctx context.Context, obj interface{}) (apiv1alpha2.Cluster, error)
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ToClusterFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterFunc must not be empty", config)
	}

	r := &Resource{
		logger:        config.Logger,
		toClusterFunc: config.ToClusterFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ensure(ctx context.Context, obj interface{}) error {
	cr, err := r.toClusterFunc(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(cr.Status.APIEndpoints) != 1 {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("cluster %#q does not have any api endpoint set in the cr status yet", key.ClusterID(&cr)))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	cc.Status.Endpoint.Base = cr.Status.APIEndpoints[0].Host

	return nil
}
