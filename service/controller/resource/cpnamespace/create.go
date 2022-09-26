package cpnamespace

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/v5/service/controller/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	ns, err := toNamespace(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if ns != nil {
		r.logger.Debugf(ctx, "creating namespace %#q in control plane", ns.Name)

		_, err = r.k8sClient.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
		if apierrors.IsAlreadyExists(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "created namespace %#q in control plane", ns.Name)
	} else {
		r.logger.Debugf(ctx, "did not create namespace %#q in control plane", key.ClusterID(&cr))
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentNamespace, err := toNamespace(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredNamespace, err := toNamespace(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var namespaceToCreate *corev1.Namespace
	if currentNamespace == nil {
		namespaceToCreate = desiredNamespace
	}

	return namespaceToCreate, nil
}
