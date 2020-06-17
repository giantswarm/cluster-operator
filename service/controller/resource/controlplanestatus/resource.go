package controlplanestatus

import (
	"context"
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/key"
	"github.com/giantswarm/cluster-operator/service/internal/nodecount"
	"github.com/giantswarm/cluster-operator/service/internal/tenantclient"
)

const (
	Name = "controlplanestatus"
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
	cr := &infrastructurev1alpha2.G8sControlPlane{}
	{
		cp, err := key.ToG8sControlPlane(obj)
		if err != nil {
			return microerror.Mask(err)
		}

		err = r.k8sClient.CtrlClient().Get(ctx, key.ObjRefToNamespacedName(key.ObjRefFromG8sControlPlane(cp)), cr)
		if err != nil {
			return microerror.Mask(err)
		}
	}
	masterNodes, err := r.nodeCount.MasterCount(ctx, cr)
	if tenantclient.IsNotAvailable(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not getting master nodes for tenant cluster %#q", key.ClusterID(cr)))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "checking if status of control plane needs to be updated")

		replicasChanged := cr.Status.Replicas != masterNodes[cr.Labels[label.ControlPlane]].Nodes
		readyReplicasChanged := cr.Status.ReadyReplicas != masterNodes[cr.Labels[label.ControlPlane]].Ready

		if !replicasChanged && !readyReplicasChanged {
			r.logger.LogCtx(ctx, "level", "debug", "message", "status of control plane does not need to be updated")
			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "status of control plane needs to be updated")
	}

	{
		cr.Status.Replicas = masterNodes[cr.Labels[label.ControlPlane]].Nodes
		cr.Status.ReadyReplicas = masterNodes[cr.Labels[label.ControlPlane]].Ready
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating status of control plane")

		err := r.k8sClient.CtrlClient().Status().Update(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated status of control plane")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
	}

	return nil
}
