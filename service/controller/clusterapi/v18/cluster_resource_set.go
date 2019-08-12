package v18

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/clusterclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	"github.com/giantswarm/tenantcluster"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v18/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v18/key"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v18/resources/awsclusterconfig"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v18/resources/clusterid"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v18/resources/clusterstatus"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v18/resources/tenantclients"
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

	var clusterIDResource controller.Resource
	{
		c := clusterid.Config{
			CMAClient:                   config.CMAClient,
			CommonClusterStatusAccessor: &key.AWSClusterStatusAccessor{},
			G8sClient:                   config.G8sClient,
			Logger:                      config.Logger,
		}

		clusterIDResource, err = clusterid.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterstatusResource controller.Resource
	{
		c := clusterstatus.Config{
			Accessor:  &key.AWSClusterStatusAccessor{},
			CMAClient: config.CMAClient,
			G8sClient: config.G8sClient,
			Logger:    config.Logger,
		}

		clusterstatusResource, err = clusterstatus.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var awsclusterconfigResource controller.Resource
	{
		c := awsclusterconfig.Config{
			ClusterClient: config.ClusterClient,
			G8sClient:     config.G8sClient,
			Logger:        config.Logger,
		}

		awsclusterconfigResource, err = awsclusterconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tenantClientsResource controller.Resource
	{
		c := tenantclients.Config{
			CMAClient:     config.CMAClient,
			Logger:        config.Logger,
			Tenant:        config.Tenant,
			ToClusterFunc: key.ToCluster,
		}

		tenantClientsResource, err = tenantclients.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []controller.Resource{
		clusterIDResource,
		tenantClientsResource,
		clusterstatusResource,

		// TODO drop this once the resources below are all actiavted.
		awsclusterconfigResource,

		// Put encryptionKeyResource first because it executes faster than
		// certConfigResource and could introduce dependency during cluster
		// creation.
		//encryptionKeyResource,
		//certConfigResource,
		//clusterConfigMapResource,
		//kubeConfigResource,

		// Following resources manage resources in tenant clusters so they
		// should be executed last.
		//namespaceResource,
		//tillerResource,
		//chartOperatorResource,
		//configMapResource,
		//chartConfigResource,
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
		ctx = controllercontext.NewContext(ctx, controllercontext.Context{})
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
