package operatorversions

import (
	"context"
	"fmt"

	"github.com/giantswarm/clusterclient"
	"github.com/giantswarm/clusterclient/service/release/searcher"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/versionbundle"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v21/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v21/key"
)

const (
	Name = "operatorversionsv21"
)

type Config struct {
	ClusterClient *clusterclient.Client
	Logger        micrologger.Logger
}

type Resource struct {
	clusterClient *clusterclient.Client
	logger        micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.ClusterClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		clusterClient: config.ClusterClient,
		logger:        config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ensure(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
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
