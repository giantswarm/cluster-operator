package v21

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	"github.com/giantswarm/tenantcluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v20/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v20/key"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v20/resources/machinedeploymentstatus"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v20/resources/tenantclients"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v20/resources/workercount"
)

type MachineDeploymentResourceSetConfig struct {
	CMAClient clientset.Interface
	G8sClient versioned.Interface
	Logger    micrologger.Logger
	Tenant    tenantcluster.Interface
}

func NewMachineDeploymentResourceSet(config MachineDeploymentResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	var machineDeploymentStatusResource resource.Interface
	{
		c := machinedeploymentstatus.Config{
			CMAClient: config.CMAClient,
			G8sClient: config.G8sClient,
			Logger:    config.Logger,
		}

		machineDeploymentStatusResource, err = machinedeploymentstatus.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tenantClientsResource resource.Interface
	{
		c := tenantclients.Config{
			Logger:        config.Logger,
			Tenant:        config.Tenant,
			ToClusterFunc: newMachineDeploymentToClusterFunc(config.CMAClient),
		}

		tenantClientsResource, err = tenantclients.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var workerCountResource resource.Interface
	{
		c := workercount.Config{
			Logger: config.Logger,

			ToClusterFunc: newMachineDeploymentToClusterFunc(config.CMAClient),
		}

		workerCountResource, err = workercount.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		tenantClientsResource,
		workerCountResource,
		machineDeploymentStatusResource,
	}

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
		cr, err := key.ToMachineDeployment(obj)
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

func newMachineDeploymentToClusterFunc(cmaClient clientset.Interface) func(obj interface{}) (v1alpha1.Cluster, error) {
	return func(obj interface{}) (v1alpha1.Cluster, error) {
		cr, err := key.ToMachineDeployment(obj)
		if err != nil {
			return v1alpha1.Cluster{}, microerror.Mask(err)
		}

		m, err := cmaClient.ClusterV1alpha1().Clusters(cr.Namespace).Get(key.ClusterID(&cr), metav1.GetOptions{})
		if err != nil {
			return v1alpha1.Cluster{}, microerror.Mask(err)
		}

		return *m, nil
	}
}
