package cpnamespace

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v8/pkg/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/v8/pkg/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/v5/service/controller/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var ns *corev1.Namespace
	{
		r.logger.Debugf(ctx, "finding namespace %#q in control plane", key.ClusterID(&cr))

		m, err := r.k8sClient.CoreV1().Namespaces().Get(ctx, key.ClusterID(&cr), metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.Debugf(ctx, "did not find namespace %#q in control plane", key.ClusterID(&cr))
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			ns = m
			r.logger.Debugf(ctx, "found namespace %#q in control plane", key.ClusterID(&cr))
		}
	}

	// In case the namespace is already terminating we do not need to do any
	// further work. We cancel the namespace resource to prevent any further work,
	// but keep the finalizers until the namespace got finally deleted. This is to
	// prevent issues with the monitoring and alerting systems. The cluster status
	// conditions of the watched CR are used to inhibit alerts, for instance when
	// the cluster is being deleted.
	if ns != nil && ns.Status.Phase == corev1.NamespaceTerminating {
		r.logger.Debugf(ctx, "namespace is %#q", corev1.NamespaceTerminating)

		r.logger.Debugf(ctx, "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.Debugf(ctx, "keeping finalizers")
		finalizerskeptcontext.SetKept(ctx)

		return nil, nil
	}

	if ns == nil && key.IsDeleted(&cr) {
		r.logger.Debugf(ctx, "resource deletion completed")

		r.logger.Debugf(ctx, "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)

		return nil, nil
	}

	return ns, nil
}
