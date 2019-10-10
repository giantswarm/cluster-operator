package updatemachinedeployments

import (
	"context"
	"fmt"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/kubernetes/pkg/apis/core"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v21/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
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

	var node core.Node
	{
		o := metav1.ListOptions{}
		list, err := k8sClient.CoreV1().Nodes().List(o)
		if tenant.IsAPINotAvailable(err) {
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
		if len(list.Items) == 0 {
			// TODO return error handling
		}

		node = list.Items[0]
	}

	var versionLabel string
	{
		l := node.GetLabels()
		n := node.GetName()

		labelProvider := "giantswarm.io/provider"
		p, ok := l[labelProvider]
		if !ok {
			return nil, microerror.Maskf(missingLabelError, labelProvider)
		}

		labelVersion = p + "-operator.giantswarm.io/version"
		v, ok := l[labelVersion]
		if !ok {
			return nil, microerror.Maskf(missingLabelError, labelVersion)
		}
	}

	for _, md := range machineDeployments {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating machine deployment %#q for tenant cluster %#q", md.Namespace+"/"+md.Name, key.ClusterID(&cr)))

		md.Labels[versionLabel] = cr.Labels[versionLabel]
		md.Labels[label.OperatorVersion] = cr.Labels[label.OperatorVersion]
		md.Labels[label.ReleaseVersion] = cr.Labels[label.ReleaseVersion]

		_, err := r.cmaClient.ClusterV1alpha1().MachineDeployments(md.Namespace).Update(&md)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated machine deployment %#q for tenant cluster %#q", md.Namespace+"/"+md.Name, key.ClusterID(&cr)))
	}

	return nil
}
