package v19

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
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/key"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/resources/awsclusterconfig"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/resources/clusterid"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/resources/clusterstatus"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/resources/operatorversions"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/resources/tenantclients"
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

	DNSIP string
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

	var clusterStatusResource controller.Resource
	{
		c := clusterstatus.Config{
			Accessor:  &key.AWSClusterStatusAccessor{},
			CMAClient: config.CMAClient,
			G8sClient: config.G8sClient,
			Logger:    config.Logger,
		}

		clusterStatusResource, err = clusterstatus.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	//var encryptionKeyGetter secretresource.StateGetter
	//{
	//	c := encryptionkey.Config{
	//		K8sClient: config.K8sClient,
	//		Logger:    config.Logger,
	//	}
	//
	//	encryptionKeyGetter, err = encryptionkey.New(c)
	//	if err != nil {
	//		return nil, microerror.Mask(err)
	//	}
	//}
	//
	//var encryptionKeyResource controller.Resource
	//{
	//	c := secretresource.Config{
	//		K8sClient: config.K8sClient,
	//		Logger:    config.Logger,
	//
	//		Name:        encryptionkey.Name,
	//		StateGetter: encryptionKeyGetter,
	//	}
	//
	//	ops, err := secretresource.New(c)
	//	if err != nil {
	//		return nil, microerror.Mask(err)
	//	}
	//
	//	encryptionKeyResource, err = toCRUDResource(config.Logger, ops)
	//	if err != nil {
	//		return nil, microerror.Mask(err)
	//	}
	//}

	//var certConfigResource controller.Resource
	//{
	//  c := certconfig.Config{
	//  	G8sClient: config.K8sClient,
	//  	K8sClient: config.K8sClient,
	//  	Logger:    config.Logger,
	//
	//  	APIIP:    config.APIIP,
	//  	CertTTL:  config.CertTTL,
	//  	Provider: config.Provider,
	//  }
	//
	//	ops, err := certconfig.New(c)
	//	if err != nil {
	//		return nil, microerror.Mask(err)
	//	}
	//
	//	certConfigResource, err = toCRUDResource(config.Logger, ops)
	//	if err != nil {
	//		return nil, microerror.Mask(err)
	//	}
	//}

	//var chartOperatorResource controller.Resource
	//{
	//	c := certconfig.Config{
	//		ApprClient: config.ApprClient,
	//		FileSystem: config.FileSystem,
	//		Logger:     config.Logger,
	//
	//		DNSIP:          config.DNSIP,
	//		RegistryDomain: config.RegistryDomain,
	//	}
	//
	//	ops, err := certconfig.New(c)
	//	if err != nil {
	//		return nil, microerror.Mask(err)
	//	}
	//
	//	chartOperatorResource, err = toCRUDResource(config.Logger, ops)
	//	if err != nil {
	//		return nil, microerror.Mask(err)
	//	}
	//}

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

	//var namespaceResource controller.Resource
	//{
	//	c := namespace.Config{
	//		Logger: config.Logger,
	//	}
	//
	//	ops, err := namespace.New(c)
	//	if err != nil {
	//		return nil, microerror.Mask(err)
	//	}
	//
	//	namespaceResource, err = toCRUDResource(config.Logger, ops)
	//	if err != nil {
	//		return nil, microerror.Mask(err)
	//	}
	//}

	var operatorVersionsResource controller.Resource
	{
		c := operatorversions.Config{
			ClusterClient: config.ClusterClient,
			Logger:        config.Logger,
		}

		operatorVersionsResource, err = operatorversions.New(c)
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
		operatorVersionsResource,
		tenantClientsResource,
		clusterStatusResource,

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

func toCRUDResource(logger micrologger.Logger, ops controller.CRUDResourceOps) (*controller.CRUDResource, error) {
	c := controller.CRUDResourceConfig{
		Logger: logger,
		Ops:    ops,
	}

	r, err := controller.NewCRUDResource(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return r, nil
}
