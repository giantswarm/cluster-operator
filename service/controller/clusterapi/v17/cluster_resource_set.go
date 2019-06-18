package v17

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/clusterclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/resource/metricsresource"
	"github.com/giantswarm/operatorkit/controller/resource/retryresource"
	"github.com/giantswarm/tenantcluster"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v17/key"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v17/resources/awsclusterconfig"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v17/resources/clusterstatus"
)

// ClusterResourceSetConfig contains necessary dependencies and settings for
// Cluster API's Cluster controller ResourceSet configuration.
type ClusterResourceSetConfig struct {
	BaseClusterConfig *cluster.Config
	ClusterClient     *clusterclient.Client
	CMAClient         clientset.Interface
	G8sClient         versioned.Interface
	Logger            micrologger.Logger
	Tenant            tenantcluster.Interface
}

// NewClusterResourceSet returns a configured Cluster API's Cluster controller
// ResourceSet.
func NewClusterResourceSet(config ClusterResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	var clusterstatusResource controller.Resource
	{
		c := clusterstatus.Config{
			CMAClient: config.CMAClient,
			G8sClient: config.G8sClient,
			Logger:    config.Logger,
			Tenant:    config.Tenant,
		}

		clusterstatusResource, err = clusterstatus.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var awsclusterconfigResource controller.Resource
	{
		c := awsclusterconfig.Config{
			BaseClusterConfig: *config.BaseClusterConfig,
			ClusterClient:     config.ClusterClient,
			CMAClient:         config.CMAClient,
			G8sClient:         config.G8sClient,
			Logger:            config.Logger,
		}

		awsclusterconfigResource, err = awsclusterconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []controller.Resource{
		clusterstatusResource,
		awsclusterconfigResource,
	}

	// Wrap resources with retry and metrics.
	{
		c := retryresource.WrapConfig{
			Logger: config.Logger,
		}

		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.WrapConfig{}
		resources, err = metricsresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		return ctx, nil
	}

	handlesFunc := func(obj interface{}) bool {
		cr, err := key.ToCluster(obj)
		if err != nil {
			return false
		}

		if key.OperatorVersion(&cr) == VersionBundle().Version {
			return true
		}

		return false
	}

	var resourceSet *controller.ResourceSet
	{
		c := controller.ResourceSetConfig{
			Handles:   handlesFunc,
			InitCtx:   initCtxFunc,
			Logger:    config.Logger,
			Resources: resources,
		}

		resourceSet, err = controller.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resourceSet, nil
}
