package operatorversions

import (
	"context"
	"fmt"

	"github.com/giantswarm/clusterclient"
	"github.com/giantswarm/clusterclient/service/release/searcher"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/versionbundle"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

const (
	Name = "operatorversions"
)

type Config struct {
	ClusterClient *clusterclient.Client
	Logger        micrologger.Logger

	ToClusterFunc func(ctx context.Context, obj interface{}) (apiv1alpha2.Cluster, error)
}

type Resource struct {
	clusterClient *clusterclient.Client
	logger        micrologger.Logger

	toClusterFunc func(ctx context.Context, obj interface{}) (apiv1alpha2.Cluster, error)
}

func New(config Config) (*Resource, error) {
	if config.ClusterClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ToClusterFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterFunc must not be empty", config)
	}

	r := &Resource{
		clusterClient: config.ClusterClient,
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

	var versionBundles []versionbundle.Bundle
	{
		req := searcher.Request{
			ReleaseVersion: key.ReleaseVersion(&cr),
		}

		res, err := r.clusterClient.Release.Searcher.Search(ctx, req)
		if err != nil {
			return microerror.Mask(err)
		}

		versionBundles = res.VersionBundles
	}

	{
		if cc.Status.Versions == nil {
			cc.Status.Versions = map[string]string{}
		}
		for _, b := range versionBundles {
			cc.Status.Versions[fmt.Sprintf("%s.giantswarm.io/version", b.Name)] = b.Version
		}
	}

	return nil
}
