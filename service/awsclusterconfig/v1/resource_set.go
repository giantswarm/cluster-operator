package v1

import (
	"context"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/resource/metricsresource"
	"github.com/giantswarm/operatorkit/framework/resource/retryresource"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/pkg/v1/resource/encryptionkey"
	"github.com/giantswarm/cluster-operator/service/awsclusterconfig/v1/key"
	"github.com/giantswarm/cluster-operator/service/awsclusterconfig/v1/resource/awsconfig"
)

const (
	// ResourceRetries presents number of retries for failed Resource
	// operation before giving up.
	ResourceRetries uint64 = 3
)

// ResourceSetConfig contains necessary dependencies and settings for
// AWSClusterConfig framework ResourceSet configuration.
type ResourceSetConfig struct {
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	HandledVersionBundles []string
	ProjectName           string
}

// NewResourceSet returns a configured AWSClusterConfig framework ResourceSet.
func NewResourceSet(config ResourceSetConfig) (*framework.ResourceSet, error) {
	var err error

	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}

	var encryptionKeyResource framework.Resource
	{
		c := encryptionkey.Config{
			K8sClient:                config.K8sClient,
			Logger:                   config.Logger,
			ProjectName:              config.ProjectName,
			ToClusterGuestConfigFunc: toClusterGuestConfig,
		}

		ops, err := encryptionkey.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		encryptionKeyResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var awsConfigResource framework.Resource
	{
		c := awsconfig.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		ops, err := awsconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		awsConfigResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []framework.Resource{
		// Put encryptionKeyResource first because it executes faster than
		// awsConfigResource and could introduce dependency during cluster
		// creation.
		encryptionKeyResource,
		awsConfigResource,
	}

	// Wrap resources with retry and metrics.
	{
		retryWrapConfig := retryresource.WrapConfig{}

		retryWrapConfig.BackOffFactory = func() backoff.BackOff { return backoff.WithMaxTries(backoff.NewExponentialBackOff(), ResourceRetries) }
		retryWrapConfig.Logger = config.Logger

		resources, err = retryresource.Wrap(resources, retryWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		metricsWrapConfig := metricsresource.WrapConfig{}

		metricsWrapConfig.Name = config.ProjectName

		resources, err = metricsresource.Wrap(resources, metricsWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		return ctx, nil
	}

	handlesFunc := func(obj interface{}) bool {
		awsClusterConfig, err := key.ToCustomObject(obj)
		if err != nil {
			return false
		}

		if key.VersionBundleVersion(awsClusterConfig) == VersionBundle().Version {
			return true
		}

		return false
	}

	var resourceSet *framework.ResourceSet
	{
		c := framework.ResourceSetConfig{
			Handles:   handlesFunc,
			InitCtx:   initCtxFunc,
			Logger:    config.Logger,
			Resources: resources,
		}

		resourceSet, err = framework.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resourceSet, nil
}

func toClusterGuestConfig(obj interface{}) (v1alpha1.ClusterGuestConfig, error) {
	awsClusterConfig, err := key.ToCustomObject(obj)
	if err != nil {
		return v1alpha1.ClusterGuestConfig{}, microerror.Mask(err)
	}

	return key.ClusterGuestConfig(awsClusterConfig), nil
}

func toCRUDResource(logger micrologger.Logger, ops framework.CRUDResourceOps) (*framework.CRUDResource, error) {
	c := framework.CRUDResourceConfig{
		Logger: logger,
		Ops:    ops,
	}

	r, err := framework.NewCRUDResource(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return r, nil
}
