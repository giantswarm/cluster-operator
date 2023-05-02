package clusterconfigmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v8/pkg/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/v5/pkg/label"
	"github.com/giantswarm/cluster-operator/v5/pkg/project"
	"github.com/giantswarm/cluster-operator/v5/service/controller/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) ([]*corev1.ConfigMap, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// The config maps are deleted when the namespace is deleted.
	if key.IsDeleted(&cr) {
		r.logger.Debugf(ctx, "not deleting config maps for tenant cluster %#q", key.ClusterID(&cr))
		r.logger.Debugf(ctx, "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil
	}

	var configMaps []*corev1.ConfigMap
	{
		r.logger.Debugf(ctx, "finding cluster config maps in namespace %#q", key.ClusterID(&cr))

		lo := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", label.ManagedBy, project.Name()),
		}

		list, err := r.k8sClient.CoreV1().ConfigMaps(key.ClusterID(&cr)).List(ctx, lo)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		for _, item := range list.Items {
			configMaps = append(configMaps, item.DeepCopy())
		}

		r.logger.Debugf(ctx, "found %d config maps in namespace %#q", len(configMaps), key.ClusterID(&cr))
	}

	return configMaps, nil
}
