package chartconfig

import (
	"reflect"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/pkg/v2/guestcluster"
	"github.com/giantswarm/cluster-operator/pkg/v2/key"
)

const (
	// Name is the identifier of the resource.
	Name = "chartconfigv1"
)

// Config represents the configuration used to create a new chart config resource.
type Config struct {
	BaseClusterConfig        cluster.Config
	G8sClient                versioned.Interface
	Guest                    guestcluster.Interface
	K8sClient                kubernetes.Interface
	Logger                   micrologger.Logger
	ProjectName              string
	ToClusterGuestConfigFunc func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
}

// Resource implements the chart config resource.
type Resource struct {
	baseClusterConfig        cluster.Config
	g8sClient                versioned.Interface
	guest                    guestcluster.Interface
	k8sClient                kubernetes.Interface
	logger                   micrologger.Logger
	projectName              string
	toClusterGuestConfigFunc func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
}

// New creates a new configured chart config resource.
func New(config Config) (*Resource, error) {
	if reflect.DeepEqual(config.BaseClusterConfig, cluster.Config{}) {
		return nil, microerror.Maskf(invalidConfigError, "%T.BaseClusterConfig must not be empty", config)
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
		baseClusterConfig:        config.BaseClusterConfig,
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

func containsChartConfig(list []*v1alpha1.ChartConfig, item *v1alpha1.ChartConfig) bool {
	for _, l := range list {
		if reflect.DeepEqual(item, l) {
			return true
		}
	}

	return false
}

func getChartConfigByName(list []*v1alpha1.ChartConfig, name string) (*v1alpha1.ChartConfig, error) {
	for _, l := range list {
		if l.Name == name {
			return l, nil
		}
	}

	return nil, microerror.Mask(notFoundError)
}

func isChartConfigModified(a, b *v1alpha1.ChartConfig) bool {
	// If the Spec section has changed we need to update.
	if !reflect.DeepEqual(a.Spec, b.Spec) {
		return true
	}
	// If the Labels have changed we also need to update.
	if !reflect.DeepEqual(a.Labels, b.Labels) {
		return true
	}

	return false
}

func toChartConfigs(v interface{}) ([]*v1alpha1.ChartConfig, error) {
	if v == nil {
		return nil, nil
	}

	chartConfigs, ok := v.([]*v1alpha1.ChartConfig)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []*v1alpha1.ChartConfig{}, v)
	}

	return chartConfigs, nil
}
