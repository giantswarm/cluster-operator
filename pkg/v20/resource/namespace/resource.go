package namespace

import (
	"context"
	"reflect"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/pkg/v19/key"
)

const (
	// Name is the identifier of the resource.
	Name = "namespacev19"

	namespaceName = "giantswarm"
)

// Config represents the configuration used to create a new namespace resource.
type Config struct {
	BaseClusterConfig        cluster.Config
	Logger                   micrologger.Logger
	Tenant                   tenantcluster.Interface
	ToClusterGuestConfigFunc func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
	ToClusterObjectMetaFunc  func(obj interface{}) (metav1.ObjectMeta, error)
}

// Resource implements the namespace resource.
type Resource struct {
	baseClusterConfig        cluster.Config
	logger                   micrologger.Logger
	tenant                   tenantcluster.Interface
	toClusterGuestConfigFunc func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
	toClusterObjectMetaFunc  func(obj interface{}) (metav1.ObjectMeta, error)
}

// New creates a new configured namespace resource.
func New(config Config) (*Resource, error) {
	if reflect.DeepEqual(config.BaseClusterConfig, cluster.Config{}) {
		return nil, microerror.Maskf(invalidConfigError, "%T.BaseClusterConfig must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ToClusterGuestConfigFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterGuestConfigFunc must not be empty", config)
	}
	if config.Tenant == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Tenant must not be empty", config)
	}
	if config.ToClusterObjectMetaFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterObjectMetaFunc must not be empty", config)
	}

	r := &Resource{
		// Dependencies.
		baseClusterConfig:        config.BaseClusterConfig,
		logger:                   config.Logger,
		tenant:                   config.Tenant,
		toClusterGuestConfigFunc: config.ToClusterGuestConfigFunc,
		toClusterObjectMetaFunc:  config.ToClusterObjectMetaFunc,
	}

	return r, nil
}

// Name returns name of the Resource.
func (r *Resource) Name() string {
	return Name
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

	guestAPIDomain, err := key.APIDomain(clusterGuestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	tenantK8sClient, err := r.tenant.NewK8sClient(ctx, clusterConfig.ClusterID, guestAPIDomain)
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

func toNamespace(v interface{}) (*corev1.Namespace, error) {
	if v == nil {
		return nil, nil
	}

	namespace, ok := v.(*corev1.Namespace)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &corev1.Namespace{}, v)
	}

	return namespace, nil
}
