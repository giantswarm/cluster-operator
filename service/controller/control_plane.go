package controller

import (
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/cluster-operator/pkg/project"
)

// ControlPlaneConfig contains necessary dependencies and settings for the
// ControlPlane controller implementation.
type ControlPlaneConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

type ControlPlane struct {
	*controller.Controller
}

func NewControlPlane(config ControlPlaneConfig) (*ControlPlane, error) {
	var err error

	var resourceSet *controller.ResourceSet
	{
		c := clusterResourceSetConfig{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		resourceSet, err = newClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var controlPlaneController *controller.Controller
	{
		c := controller.Config{
			CRD:       infrastructurev1alpha2.NewG8sControlPlaneCRD(),
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			ResourceSets: []*controller.ResourceSet{
				resourceSet,
			},
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(infrastructurev1alpha2.G8sControlPlane)
			},

			// Name is used to compute finalizer names. This here results in something
			// like operatorkit.giantswarm.io/cluster-operator-control-plane-controller.
			Name: project.Name() + "-control-plane-controller",
		}

		controlPlaneController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &ControlPlane{
		Controller: controlPlaneController,
	}

	return c, nil
}
