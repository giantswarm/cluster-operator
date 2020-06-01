package machinedeploymentstatus

import (
	"context"

	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

const (
	Name = "machinedeploymentstatus"
)

type Config struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

type Resource struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ensure(ctx context.Context, obj interface{}) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	cr := &apiv1alpha2.MachineDeployment{}
	{
		md, err := key.ToMachineDeployment(obj)
		if err != nil {
			return microerror.Mask(err)
		}

		err = r.k8sClient.CtrlClient().Get(ctx, key.ObjRefToNamespacedName(key.ObjRefFromMachineDeployment(md)), cr)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "checking if status of machine deployment needs to be updated")

		replicasChanged := cr.Status.Replicas != cc.Status.Worker[cr.Labels[label.MachineDeployment]].Nodes
		readyReplicasChanged := cr.Status.ReadyReplicas != cc.Status.Worker[cr.Labels[label.MachineDeployment]].Ready

		if !replicasChanged && !readyReplicasChanged {
			r.logger.LogCtx(ctx, "level", "debug", "message", "status of machine deployment does not need to be updated")
			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "status of machine deployment needs to be updated")
	}

	{
		cr.Status.Replicas = cc.Status.Worker[cr.Labels[label.MachineDeployment]].Nodes
		cr.Status.ReadyReplicas = cc.Status.Worker[cr.Labels[label.MachineDeployment]].Ready
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating status of machine deployment")

		err := r.k8sClient.CtrlClient().Status().Update(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated status of machine deployment")

		if key.IsDeleted(cr) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "keeping finalizers")
			finalizerskeptcontext.SetKept(ctx)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
	}

	return nil
}
