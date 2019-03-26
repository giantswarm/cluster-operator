package namespace

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	clusterGuestConfig, err := r.toClusterGuestConfigFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	namespace, err := r.desiredNamespace(ctx, clusterGuestConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating %#q namespace in tenant cluster", namespace.Name))

	tenantK8sClient, err := r.getTenantK8sClient(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = tenantK8sClient.CoreV1().Namespaces().Create(namespace)
	if apierrors.IsAlreadyExists(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("%#q namespace already created in tenant cluster", namespace.Name))

		return nil
	} else if apierrors.IsTimeout(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster api timeout.")

		// We should not hammer tenant API if it is not available, the tenant cluster
		// might be initializing. We will retry on next reconciliation loop.
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	} else if tenant.IsAPINotAvailable(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is not available.")

		// We should not hammer tenant API if it is not available, the tenant cluster
		// might be initializing. We will retry on next reconciliation loop.
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created %#q namespace in tenant cluster", namespace.Name))

	return nil
}

func (r *Resource) desiredNamespace(ctx context.Context, clusterGuestConfig v1alpha1.ClusterGuestConfig) (*corev1.Namespace, error) {
	clusterConfig, err := prepareClusterConfig(r.baseClusterConfig, clusterGuestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	namespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
			Labels: map[string]string{
				label.Cluster:      clusterConfig.ClusterID,
				label.ManagedBy:    r.projectName,
				label.Organization: clusterConfig.Organization,
			},
		},
	}

	return namespace, nil
}
