package chart

import (
	"reflect"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/pkg/v1/guestcluster"
	"github.com/giantswarm/cluster-operator/pkg/v1/key"
)

const (
	// Name is the identifier of the resource.
	Name = "chartv1"

	chartOperatorChart   = "chart-operator-chart"
	chartOperatorChannel = "0-1-beta"
	chartOperatorRelease = "chart-operator"
)

// Config represents the configuration used to create a new chart config resource.
type Config struct {
	ApprClient               apprclient.Interface
	BaseClusterConfig        cluster.Config
	Fs                       afero.Fs
	G8sClient                versioned.Interface
	Guest                    guestcluster.Interface
	K8sClient                kubernetes.Interface
	Logger                   micrologger.Logger
	ProjectName              string
	ToClusterGuestConfigFunc func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
}

// Resource implements the chart resource.
type Resource struct {
	apprClient               apprclient.Interface
	baseClusterConfig        cluster.Config
	fs                       afero.Fs
	g8sClient                versioned.Interface
	guest                    guestcluster.Interface
	k8sClient                kubernetes.Interface
	logger                   micrologger.Logger
	projectName              string
	toClusterGuestConfigFunc func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
}

// New creates a new configured chart resource.
func New(config Config) (*Resource, error) {
	if config.ApprClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ApprClient must not be empty", config)
	}
	if reflect.DeepEqual(config.BaseClusterConfig, cluster.Config{}) {
		return nil, microerror.Maskf(invalidConfigError, "%T.BaseClusterConfig must not be empty", config)
	}
	if config.Fs == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Fs must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Guest == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Guest must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}
	if config.ToClusterGuestConfigFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterGuestConfigFunc must not be empty", config)
	}

	newResource := &Resource{
		apprClient:               config.ApprClient,
		baseClusterConfig:        config.BaseClusterConfig,
		fs:                       config.Fs,
		g8sClient:                config.G8sClient,
		guest:                    config.Guest,
		k8sClient:                config.K8sClient,
		logger:                   config.Logger,
		projectName:              config.ProjectName,
		toClusterGuestConfigFunc: config.ToClusterGuestConfigFunc,
	}

	return newResource, nil
}

// Name returns name of the Resource.
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

	clusterConfig.Organization = clusterGuestConfig.Owner

	return clusterConfig, nil
}

func toResourceState(v interface{}) (ResourceState, error) {
	if v == nil {
		return ResourceState{}, nil
	}

	resourceState, ok := v.(*ResourceState)
	if !ok {
		return ResourceState{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", resourceState, v)
	}

	return *resourceState, nil
}
