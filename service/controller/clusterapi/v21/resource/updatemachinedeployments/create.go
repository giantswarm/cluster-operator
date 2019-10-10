package updatemachinedeployments

import (
	"context"
	"fmt"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v21/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v21/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	if cc.Client.TenantCluster.K8s == nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster clients not available in controller context")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	}

	var machineDeployments []v1alpha1.MachineDeployment
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding machine deployments for tenant cluster")

		l := metav1.AddLabelToSelector(
			&metav1.LabelSelector{},
			label.Cluster,
			key.ClusterID(&cr),
		)
		o := metav1.ListOptions{
			LabelSelector: labels.Set(l.MatchLabels).String(),
		}

		list, err := r.cmaClient.ClusterV1alpha1().MachineDeployments(cr.Namespace).List(o)
		if err != nil {
			return microerror.Mask(err)
		}

		machineDeployments = list.Items

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d machine deployments for tenant cluster", len(machineDeployments)))
	}

	var node corev1.Node
	{
		o := metav1.ListOptions{}
		list, err := cc.Client.TenantCluster.K8s.CoreV1().Nodes().List(o)
		if tenant.IsAPINotAvailable(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		if len(list.Items) == 0 {
			r.logger.LogCtx(ctx, "level", "debug", "message", "no tenant cluster nodes available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		node = list.Items[0]
	}

	// versionLabel contains computed provider operator version label which
	// contains the operator's version bundle version. Here we need to dynamically
	// compute it based on the provider we are running in. This approach is based
	// on the provider label tracked within the Tenant Cluster nodes.
	var versionLabel string
	{
		p, ok := node.Labels[label.Provider]
		if !ok {
			return microerror.Maskf(missingLabelError, label.Provider)
		}

		versionLabel = p + "-operator.giantswarm.io/version"
	}

	for _, md := range machineDeployments {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating machine deployment %#q for tenant cluster %#q", md.Namespace+"/"+md.Name, key.ClusterID(&cr)))

		var updated bool

		// Syncing the Provider Operator version. For instance aws-operator,
		// kvm-operator or the like.
		{
			l := versionLabel
			d, ok := cr.Labels[l]
			c := md.Labels[l]
			if ok && d != "" && d != md.Labels[l] {
				md.Labels[l] = d
				updated = true

				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("label value of %#q changed from %#q to %#q", l, c, d))
			}
		}

		// Syncing the cluster-operator version.
		{
			l := label.OperatorVersion
			d, ok := cr.Labels[l]
			c := md.Labels[l]
			if ok && d != "" && d != md.Labels[l] {
				md.Labels[l] = d
				updated = true

				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("label value of %#q changed from %#q to %#q", l, c, d))
			}
		}

		// Syncing the Giant Swarm Release version.
		{
			l := label.ReleaseVersion
			d, ok := cr.Labels[l]
			c := md.Labels[l]
			if ok && d != "" && d != md.Labels[l] {
				md.Labels[l] = d
				updated = true

				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("label value of %#q changed from %#q to %#q", l, c, d))
			}
		}

		if updated {
			_, err := r.cmaClient.ClusterV1alpha1().MachineDeployments(md.Namespace).Update(&md)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated machine deployment %#q for tenant cluster %#q", md.Namespace+"/"+md.Name, key.ClusterID(&cr)))
	}

	return nil
}
