package chartoperator

import (
	"context"
	"reflect"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster"
	"github.com/spf13/afero"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/pkg/v20/key"
)

const (
	// Name is the identifier of the resource.
	Name = "chartoperatorv20"

	chartOperatorChart         = "chart-operator-chart"
	chartOperatorChannel       = "0-9-stable"
	chartOperatorDeployment    = "chart-operator"
	chartOperatorRelease       = "chart-operator"
	chartOperatorNamespace     = "giantswarm"
	chartOperatorDesiredStatus = "DEPLOYED"
	chartOperatorFailedStatus  = "FAILED"
)

// Config represents the configuration used to create a new chartoperator resource.
type Config struct {
	ApprClient               apprclient.Interface
	BaseClusterConfig        cluster.Config
	ClusterIPRange           string
	Fs                       afero.Fs
	G8sClient                versioned.Interface
	K8sClient                kubernetes.Interface
	Logger                   micrologger.Logger
	ProjectName              string
	RegistryDomain           string
	Tenant                   tenantcluster.Interface
	ToClusterGuestConfigFunc func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
	ToClusterObjectMetaFunc  func(obj interface{}) (metav1.ObjectMeta, error)
}

// Resource implements the chartoperator resource.
type Resource struct {
	apprClient               apprclient.Interface
	baseClusterConfig        cluster.Config
	clusterIPRange           string
	fs                       afero.Fs
	g8sClient                versioned.Interface
	k8sClient                kubernetes.Interface
	logger                   micrologger.Logger
	projectName              string
	registryDomain           string
	tenant                   tenantcluster.Interface
	toClusterGuestConfigFunc func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
	toClusterObjectMetaFunc  func(obj interface{}) (metav1.ObjectMeta, error)
}

// New creates a new configured chartoperator resource.
func New(config Config) (*Resource, error) {
	if config.ApprClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ApprClient must not be empty", config)
	}
	if reflect.DeepEqual(config.BaseClusterConfig, cluster.Config{}) {
		return nil, microerror.Maskf(invalidConfigError, "%T.BaseClusterConfig must not be empty", config)
	}
	if config.ClusterIPRange == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterIPRange must not be empty", config)
	}
	if config.Fs == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Fs must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
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
	if config.RegistryDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.RegistryDomain must not be empty", config)
	}
	if config.Tenant == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Tenant must not be empty", config)
	}
	if config.ToClusterGuestConfigFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterGuestConfigFunc must not be empty", config)
	}
	if config.ToClusterObjectMetaFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterObjectMetaFunc must not be empty", config)
	}

	newResource := &Resource{
		apprClient:               config.ApprClient,
		baseClusterConfig:        config.BaseClusterConfig,
		clusterIPRange:           config.ClusterIPRange,
		fs:                       config.Fs,
		g8sClient:                config.G8sClient,
		k8sClient:                config.K8sClient,
		logger:                   config.Logger,
		projectName:              config.ProjectName,
		registryDomain:           config.RegistryDomain,
		tenant:                   config.Tenant,
		toClusterGuestConfigFunc: config.ToClusterGuestConfigFunc,
		toClusterObjectMetaFunc:  config.ToClusterObjectMetaFunc,
	}

	return newResource, nil
}

// Name returns name of the Resource.
func (r *Resource) Name() string {
	return Name
}

func (r *Resource) getTenantHelmClient(ctx context.Context, obj interface{}) (helmclient.Interface, error) {
	clusterGuestConfig, err := r.toClusterGuestConfigFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterConfig, err := prepareClusterConfig(r.baseClusterConfig, clusterGuestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	guestAPIDomain, err := key.APIDomain(clusterGuestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	tenantHelmClient, err := r.tenant.NewHelmClient(ctx, clusterConfig.ClusterID, guestAPIDomain)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return tenantHelmClient, nil
}

func (r *Resource) getTenantK8sClient(ctx context.Context, obj interface{}) (kubernetes.Interface, error) {
	clusterGuestConfig, err := r.toClusterGuestConfigFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterConfig, err := prepareClusterConfig(r.baseClusterConfig, clusterGuestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	tenantAPIDomain, err := key.APIDomain(clusterGuestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	tenantK8sClient, err := r.tenant.NewK8sClient(ctx, clusterConfig.ClusterID, tenantAPIDomain)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return tenantK8sClient, nil
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

func shouldUpdate(currentState, desiredState ResourceState) bool {
	if currentState.ReleaseVersion != "" && currentState.ReleaseVersion != desiredState.ReleaseVersion {
		// ReleaseVersion has changed for the channel so we need to update the Helm
		// Release.
		return true
	}

	if !reflect.DeepEqual(currentState.ChartValues, desiredState.ChartValues) {
		return true
	}

	if currentState.ReleaseStatus == chartOperatorFailedStatus {
		// Release status is failed so do force upgrade to attempt to fix it.
		return true
	}

	return false
}
