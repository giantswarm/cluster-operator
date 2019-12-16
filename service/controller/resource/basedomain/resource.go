package basedomain

import (
	"context"
	"fmt"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/key"
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

	// In order to get the base domain from the API Endpoint we need to remove the
	// API relevent prefix. The CR status contains an API Endpoint like the
	// following.
	//
	//     api.n2fm4.k8s.gauss.eu-central-1.aws.gigantic.io
	//
	// What we want to dispatch via the controller context is something like this.
	//
	//     gauss.eu-central-1.aws.gigantic.io
	//
	cc.Status.Endpoint.Base = strings.Replace(cr.Status.APIEndpoints[0].Host, fmt.Sprintf("api.%s.k8s.", key.ClusterID(&cr)), "", 1)

	return nil
}
