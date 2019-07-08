package machinedeploymentstatus

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v18/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v18/key"
)

const (
	Name = "machinedeploymentstatusv18"
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
	var err error

	cr, err := key.ToMachineDeployment(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var nodes []corev1.Node
	var ready []corev1.Node
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding nodes of tenant cluster")

		if cc.Client.TenantCluster.K8s == nil {
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster k8s client is not initialized")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		}

		o := metav1.ListOptions{
			// This label selector excludes the master nodes from node list.
			//
			// Constructing this LabelSelector is not currently possible with
			// k8s types and functions. Therefore it's hardcoded here.
			LabelSelector: fmt.Sprintf("!%s", label.MasterNodeRole),
		}

		l, err := cc.Client.TenantCluster.K8s.CoreV1().Nodes().List(o)
		if tenant.IsAPINotAvailable(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant API not available", "stack", microerror.Stack(err))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		nodes = l.Items

		for _, n := range nodes {
			for _, c := range n.Status.Conditions {
				if c.Type == corev1.NodeReady && c.Status == corev1.ConditionTrue {
					ready = append(ready, n)
				}
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found nodes of tenant cluster")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "checking if status of machine deployment needs to be updated")

		replicasChanged := cr.Status.Replicas != int32(len(nodes))
		readyReplicasChanged := cr.Status.ReadyReplicas != int32(len(ready))

		if !replicasChanged && !readyReplicasChanged {
			r.logger.LogCtx(ctx, "level", "debug", "message", "status of machine deployment does not need to be updated")
			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "status of machine deployment needs to be updated")
	}

	var md *v1alpha1.MachineDeployment
	{
		md, err = r.cmaClient.ClusterV1alpha1().MachineDeployments(cr.Namespace).Get(cr.Name, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		md.Status.Replicas = int32(len(nodes))
		md.Status.ReadyReplicas = int32(len(ready))
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
