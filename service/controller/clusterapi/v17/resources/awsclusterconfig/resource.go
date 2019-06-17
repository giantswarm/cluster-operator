package awsclusterconfig

import (
	"reflect"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/clusterclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/pkg/v6/key"
)

const (
	Name = "awsclusterconfigv17"

	// With first version of Node Pools implementation, the maximum number of
	// AZs for a tenant cluster is always 4. This is due to restrictions in
	// current network design.
	NumberOfAZsWithNodePools = 4
)

// Config represents the configuration used to create a new awsclusterconfig resource.
type Config struct {
	BaseClusterConfig cluster.Config
	ClusterClient     *clusterclient.Client
	CMAClient         clientset.Interface
	G8sClient         versioned.Interface
	Logger            micrologger.Logger
}

// Resource implements the awsclusterconfig resource.
type Resource struct {
	baseClusterConfig cluster.Config
	clusterClient     *clusterclient.Client
	cmaClient         clientset.Interface
	g8sClient         versioned.Interface
	logger            micrologger.Logger
}

// New creates a new configured tiller resource.
func New(config Config) (*Resource, error) {
	if reflect.DeepEqual(config.BaseClusterConfig, cluster.Config{}) {
		return nil, microerror.Maskf(invalidConfigError, "%T.BaseClusterConfig must not be empty", config)
	}
	if config.ClusterClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterClient must not be empty", config)
	}
	if config.CMAClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CMAClient must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		baseClusterConfig: config.BaseClusterConfig,
		clusterClient:     config.ClusterClient,
		cmaClient:         config.CMAClient,
		g8sClient:         config.G8sClient,
		logger:            config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func prepareClusterConfig(baseClusterConfig cluster.Config, clusterGuestConfig v1alpha1.ClusterGuestConfig) (cluster.Config, error) {
	var err error

	// Use baseClusterConfig as a basis and supplement it with information from
	// clusterGuestConfig.
	clusterConfig := baseClusterConfig

	clusterConfig.ClusterID = key.ClusterID(clusterGuestConfig)
	clusterConfig.Domain.API, err = key.APIDomain(clusterGuestConfig)
	if err != nil {
		return cluster.Config{}, microerror.Mask(err)
	}

	return clusterConfig, nil
}
