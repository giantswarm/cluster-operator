package controller

import (
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v7/pkg/controller"
	"github.com/giantswarm/operatorkit/v7/pkg/resource"
	"github.com/giantswarm/operatorkit/v7/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v7/pkg/resource/wrapper/retryresource"
	"github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	apiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/v3/pkg/label"
	"github.com/giantswarm/cluster-operator/v3/pkg/project"
	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
	"github.com/giantswarm/cluster-operator/v3/service/controller/resource/deleteinfrarefs"
	"github.com/giantswarm/cluster-operator/v3/service/controller/resource/keepforinfrarefs"
	"github.com/giantswarm/cluster-operator/v3/service/controller/resource/machinedeploymentstatus"
	"github.com/giantswarm/cluster-operator/v3/service/controller/resource/updateinfrarefs"
	"github.com/giantswarm/cluster-operator/v3/service/internal/basedomain"
	"github.com/giantswarm/cluster-operator/v3/service/internal/nodecount"
	"github.com/giantswarm/cluster-operator/v3/service/internal/recorder"
	"github.com/giantswarm/cluster-operator/v3/service/internal/releaseversion"
)

type MachineDeploymentConfig struct {
	BaseDomain     basedomain.Interface
	Event          recorder.Interface
	K8sClient      k8sclient.Interface
	Logger         micrologger.Logger
	NodeCount      nodecount.Interface
	Tenant         tenantcluster.Interface
	ReleaseVersion releaseversion.Interface

	Provider string
}

type MachineDeployment struct {
	*controller.Controller
}

func NewMachineDeployment(config MachineDeploymentConfig) (*MachineDeployment, error) {
	var err error

	var resources []resource.Interface
	{
		resources, err = newMachineDeploymentResources(config)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterController *controller.Controller
	{
		c := controller.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			NewRuntimeObjectFunc: func() ctrlClient.Object {
				return new(apiv1beta1.MachineDeployment)
			},
			Resources: resources,

			// Name is used to compute finalizer names. This here results in something
			// like operatorkit.giantswarm.io/cluster-operator-machine-deployment-controller.
			Name: project.Name() + "-machine-deployment-controller",
			Selector: labels.SelectorFromSet(map[string]string{
				label.OperatorVersion: project.Version(),
			}),
		}

		clusterController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &MachineDeployment{
		Controller: clusterController,
	}

	return c, nil
}

func newMachineDeploymentResources(config MachineDeploymentConfig) ([]resource.Interface, error) {
	var err error

	var deleteInfraRefsResource resource.Interface
	{
		c := deleteinfrarefs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			Provider: config.Provider,
			ToObjRef: toMachineDeploymentObjRef,
		}

		deleteInfraRefsResource, err = deleteinfrarefs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var keepForInfraRefsResource resource.Interface
	{
		c := keepforinfrarefs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			ToObjRef: toMachineDeploymentObjRef,
		}

		keepForInfraRefsResource, err = keepforinfrarefs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var machineDeploymentStatusResource resource.Interface
	{
		c := machinedeploymentstatus.Config{
			Event:     config.Event,
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			NodeCount: config.NodeCount,
		}

		machineDeploymentStatusResource, err = machinedeploymentstatus.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var updateInfraRefsResource resource.Interface
	{
		c := updateinfrarefs.Config{
			K8sClient:      config.K8sClient,
			Logger:         config.Logger,
			ReleaseVersion: config.ReleaseVersion,

			ToObjRef: toMachineDeploymentObjRef,
			Provider: config.Provider,
		}

		updateInfraRefsResource, err = updateinfrarefs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		// Following resources manage CR status information. Note that
		// keepForInfraRefsResource needs to run before
		// machineDeploymentStatusResource because keepForInfraRefsResource keeps
		// finalizers where machineDeploymentStatusResource does not.
		machineDeploymentStatusResource,

		// Following resources manage resources in the control plane.
		deleteInfraRefsResource,
		keepForInfraRefsResource,
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

	return resources, nil
}

func toMachineDeploymentObjRef(obj interface{}) (corev1.ObjectReference, error) {
	cr, err := key.ToMachineDeployment(obj)
	if err != nil {
		return corev1.ObjectReference{}, microerror.Mask(err)
	}

	return key.ObjRefFromMachineDeployment(cr), nil
}
