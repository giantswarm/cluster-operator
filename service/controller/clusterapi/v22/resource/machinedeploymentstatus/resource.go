package machinedeploymentstatus

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v22/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v22/key"
)

const (
	Name = "machinedeploymentstatusv22"
)

type Config struct {
	CMAClient clientset.Interface
	G8sClient versioned.Interface
	Logger    micrologger.Logger
}

type Resource struct {
	cmaClient clientset.Interface
	g8sClient versioned.Interface
	logger    micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.CMAClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CMAClient must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		cmaClient: config.CMAClient,
		g8sClient: config.G8sClient,
		logger:    config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ensure(ctx context.Context, obj interface{}) error {
	cr, err := key.ToMachineDeployment(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var md *v1alpha1.MachineDeployment
	{
		md, err = r.cmaClient.ClusterV1alpha1().MachineDeployments(cr.Namespace).Get(cr.Name, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "checking if status of machine deployment needs to be updated")

		replicasChanged := cr.Status.Replicas != cc.Status.Worker[md.Labels[label.MachineDeployment]].Nodes
		readyReplicasChanged := cr.Status.ReadyReplicas != cc.Status.Worker[md.Labels[label.MachineDeployment]].Ready

		if !replicasChanged && !readyReplicasChanged {
			r.logger.LogCtx(ctx, "level", "debug", "message", "status of machine deployment does not need to be updated")
			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "status of machine deployment needs to be updated")
	}

	{
		md.Status.Replicas = cc.Status.Worker[md.Labels[label.MachineDeployment]].Nodes
		md.Status.ReadyReplicas = cc.Status.Worker[md.Labels[label.MachineDeployment]].Ready
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating status of machine deployment")

		_, err := r.cmaClient.ClusterV1alpha1().MachineDeployments(md.Namespace).UpdateStatus(md)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated status of machine deployment")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
	}

	return nil
}
