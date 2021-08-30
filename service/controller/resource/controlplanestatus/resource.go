package controlplanestatus

import (
	"context"
	"fmt"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v5/pkg/controller/context/reconciliationcanceledcontext"
	"github.com/giantswarm/operatorkit/v5/pkg/controller/context/resourcecanceledcontext"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/cluster-operator/v3/pkg/label"
	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
	"github.com/giantswarm/cluster-operator/v3/service/internal/basedomain"
	"github.com/giantswarm/cluster-operator/v3/service/internal/nodecount"
	"github.com/giantswarm/cluster-operator/v3/service/internal/recorder"
	"github.com/giantswarm/cluster-operator/v3/service/internal/tenantclient"
)

const (
	Name = "controlplanestatus"
)

type Config struct {
	Event     recorder.Interface
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
	NodeCount nodecount.Interface
}

type Resource struct {
	event     recorder.Interface
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
	nodeCount nodecount.Interface
}

func New(config Config) (*Resource, error) {
	if config.Event == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Event must not be empty", config)
	}
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
		event:     config.Event,
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
	cr := &infrastructurev1alpha3.G8sControlPlane{}
	{
		cp, err := key.ToG8sControlPlane(obj)
		if err != nil {
			return microerror.Mask(err)
		}

		err = r.k8sClient.CtrlClient().Get(ctx, types.NamespacedName{Name: cp.Name, Namespace: cp.Namespace}, cr)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	masterNodes, err := r.nodeCount.MasterCount(ctx, cr)
	if tenantclient.IsNotAvailable(err) {
		r.logger.LogCtx(
			ctx,
			"level", "debug",
			"message", fmt.Sprintf("not getting master nodes for tenant cluster %#q", key.ClusterID(cr)),
			"reason", "tenant cluster api not available yet",
		)
		r.logger.Debugf(ctx, "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil
	} else if basedomain.IsNotFound(err) {
		// in case of a cluster deletion AWSCluster CR does not exist anymore, handle basedomain error gracefully
		r.logger.Debugf(ctx, "not getting basedomain for tenant cluster %#q", key.ClusterID(cr))
		r.logger.Debugf(ctx, "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	{
		r.logger.Debugf(ctx, "checking if status of control plane needs to be updated")

		replicasChanged := cr.Status.Replicas != masterNodes[cr.Labels[label.ControlPlane]].Nodes
		readyReplicasChanged := cr.Status.ReadyReplicas != masterNodes[cr.Labels[label.ControlPlane]].Ready

		if !replicasChanged && !readyReplicasChanged {
			r.logger.Debugf(ctx, "status of control plane does not need to be updated")
			return nil
		}

		r.logger.Debugf(ctx, "status of control plane needs to be updated")
	}

	{
		cr.Status.Replicas = masterNodes[cr.Labels[label.ControlPlane]].Nodes
		cr.Status.ReadyReplicas = masterNodes[cr.Labels[label.ControlPlane]].Ready
	}

	{
		r.logger.Debugf(ctx, "updating status of control plane")
		r.event.Emit(ctx, cr, "ControlPlaneUpdated",
			fmt.Sprintf("updated status of control plane, changed replicas %d -> %d", cr.Status.Replicas, cr.Status.ReadyReplicas),
		)

		err := r.k8sClient.CtrlClient().Status().Update(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "updated status of control plane")
		r.logger.Debugf(ctx, "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
	}

	return nil
}
