package v1

import (
	"context"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/cluster-operator/service/kvm/v1/key"
	"github.com/giantswarm/cluster-operator/service/kvm/v1/resource/kvmclusterconfig"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/resource/metricsresource"
	"github.com/giantswarm/operatorkit/framework/resource/retryresource"
	"k8s.io/client-go/kubernetes"
)

const (
	ResourceRetries uint64 = 3
)

type ResourceSetConfig struct {
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	HandledVersionBundles []string
	// Name is the project name.
	Name string
}

func NewResourceSet(config ResourceSetConfig) (*framework.ResourceSet, error) {
	var err error

	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	var kvmClusterConfigResource framework.Resource
	{
		c := kvmclusterconfig.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		kvmClusterConfigResource, err = kvmclusterconfig.New(c)
		if err != nil {
			return nil, microerror.Maskf(err, "kvmclusterconfig.New")
		}
	}

	resources := []framework.Resource{
		kvmClusterConfigResource,
	}

	// Wrap resources with retry and metrics.
	{
		retryWrapConfig := retryresource.WrapConfig{}

		retryWrapConfig.BackOffFactory = func() backoff.BackOff { return backoff.WithMaxTries(backoff.NewExponentialBackOff(), ResourceRetries) }
		retryWrapConfig.Logger = config.Logger

		resources, err = retryresource.Wrap(resources, retryWrapConfig)
		if err != nil {
			return nil, microerror.Maskf(err, "retryresource.Wrap")
		}

		metricsWrapConfig := metricsresource.WrapConfig{}

		metricsWrapConfig.Name = config.Name

		resources, err = metricsresource.Wrap(resources, metricsWrapConfig)
		if err != nil {
			return nil, microerror.Maskf(err, "metricsresource.Wrap")
		}
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		return ctx, nil
	}

	handlesFunc := func(obj interface{}) bool {
		_, err := key.ToCustomObject(obj)
		if err != nil {
			return false
		}

		// Currently there is only one version to be handled. As long as the
		// object is of right type, it's good to go.

		return true
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
			return nil, microerror.Maskf(err, "framework.NewResourceSet")
		}
	}

	return resourceSet, nil
}
