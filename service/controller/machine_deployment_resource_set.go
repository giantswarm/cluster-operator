package controller

import (
	"context"

	"github.com/giantswarm/clusterclient"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	"github.com/giantswarm/tenantcluster"
	"k8s.io/apimachinery/pkg/types"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/key"
	"github.com/giantswarm/cluster-operator/service/controller/resource/basedomain"
	"github.com/giantswarm/cluster-operator/service/controller/resource/machinedeploymentstatus"
	"github.com/giantswarm/cluster-operator/service/controller/resource/operatorversions"
	"github.com/giantswarm/cluster-operator/service/controller/resource/tenantclients"
	"github.com/giantswarm/cluster-operator/service/controller/resource/updateinfrarefs"
	"github.com/giantswarm/cluster-operator/service/controller/resource/workercount"
)

type machineDeploymentResourceSetConfig struct {
	ClusterClient *clusterclient.Client
	K8sClient     k8sclient.Interface
	Logger        micrologger.Logger
	Tenant        tenantcluster.Interface

	Provider string
}

func newMachineDeploymentResourceSet(config machineDeploymentResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	var baseDomainResource resource.Interface
	{
		c := basedomain.Config{
			Logger: config.Logger,

			ToClusterFunc: newMachineDeploymentToClusterFunc(config.K8sClient),
		}

		baseDomainResource, err = basedomain.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var machineDeploymentStatusResource resource.Interface
	{
		c := machinedeploymentstatus.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		machineDeploymentStatusResource, err = machinedeploymentstatus.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var operatorVersionsResource resource.Interface
	{
		c := operatorversions.Config{
			ClusterClient: config.ClusterClient,
			Logger:        config.Logger,

			ToClusterFunc: newMachineDeploymentToClusterFunc(config.K8sClient),
		}

		operatorVersionsResource, err = operatorversions.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tenantClientsResource resource.Interface
	{
		c := tenantclients.Config{
			Logger:        config.Logger,
			Tenant:        config.Tenant,
			ToClusterFunc: newMachineDeploymentToClusterFunc(config.K8sClient),
		}

		tenantClientsResource, err = tenantclients.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var updateInfraRefsResource resource.Interface
	{
		c := updateinfrarefs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			ToNamespacedName: toMachineDeploymentNamespacedName,
			Provider:         config.Provider,
		}

		updateInfraRefsResource, err = updateinfrarefs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var workerCountResource resource.Interface
	{
		c := workercount.Config{
			Logger: config.Logger,

			ToClusterFunc: newMachineDeploymentToClusterFunc(config.K8sClient),
		}

		workerCountResource, err = workercount.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		// Following resources manage controller context information.
		baseDomainResource,
		operatorVersionsResource,
		tenantClientsResource,
		workerCountResource,

		// Following resources manage CR status information.
		machineDeploymentStatusResource,

		// Following resources manage resources in the control plane.
		updateInfraRefsResource,
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

		if key.OperatorVersion(&cr) == project.BundleVersion() {
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

func newMachineDeploymentToClusterFunc(k8sClient k8sclient.Interface) func(ctx context.Context, obj interface{}) (apiv1alpha2.Cluster, error) {
	return func(ctx context.Context, obj interface{}) (apiv1alpha2.Cluster, error) {
		cr := &apiv1alpha2.Cluster{}
		{
			md, err := key.ToMachineDeployment(obj)
			if err != nil {
				return apiv1alpha2.Cluster{}, microerror.Mask(err)
			}

			// Note that we cannot use key.MachineDeploymentInfraRef here because we
			// do not need to fetch the Machine Deployment again. We need to lookup
			// the Cluster CR based on the MachineDeployment CR. This is why we use
			// types.NamespacedName here explicitly.
			err = k8sClient.CtrlClient().Get(ctx, types.NamespacedName{Name: key.ClusterID(&md), Namespace: md.Namespace}, cr)
			if err != nil {
				return apiv1alpha2.Cluster{}, microerror.Mask(err)
			}
		}

		return *cr, nil
	}
}

func toMachineDeploymentNamespacedName(obj interface{}) (types.NamespacedName, error) {
	cr, err := key.ToMachineDeployment(obj)
	if err != nil {
		return types.NamespacedName{}, microerror.Mask(err)
	}

	return key.MachineDeploymentInfraRef(cr), nil
}
