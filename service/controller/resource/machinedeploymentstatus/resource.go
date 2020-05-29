package machinedeploymentstatus

import (
	"context"
	"fmt"

	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/key"
	"github.com/giantswarm/cluster-operator/service/internal/nodecount"
)

const (
	Name = "machinedeploymentstatus"
)

type Config struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
	NodeCount nodecount.Interface
}

type Resource struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
	nodeCount nodecount.Interface
}

func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.NodeCount == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.NodeCount must not be empty", config)
	}

	r := &Resource{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
		nodeCount: config.NodeCount,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ensure(ctx context.Context, obj interface{}) error {
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
	workerCount, err := r.nodeCount.WorkerCount(ctx, cr)
	if nodecount.IsTenantClusterInitialized(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not getting worker nodes for tenant cluster %#q", key.ClusterID(cr)))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "checking if status of machine deployment needs to be updated")

		replicasChanged := cr.Status.Replicas != workerCount[cr.Labels[label.MachineDeployment]].Nodes
		readyReplicasChanged := cr.Status.ReadyReplicas != workerCount[cr.Labels[label.MachineDeployment]].Ready

		if !replicasChanged && !readyReplicasChanged {
			r.logger.LogCtx(ctx, "level", "debug", "message", "status of machine deployment does not need to be updated")
			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "status of machine deployment needs to be updated")
	}

	{
		cr.Status.Replicas = workerCount[cr.Labels[label.MachineDeployment]].Nodes
		cr.Status.ReadyReplicas = workerCount[cr.Labels[label.MachineDeployment]].Ready
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating status of machine deployment")

		err := r.k8sClient.CtrlClient().Status().Update(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated status of machine deployment")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
	}

	return nil
}
