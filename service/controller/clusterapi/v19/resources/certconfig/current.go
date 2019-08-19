package certconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/key"
)

// GetCurrentState takes observed custom object as an input and based on that
// information looks for current state of cluster certconfigs and returns them.
// Return value is of type []*v1alpha1.CertConfig.
func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var certConfigs []*v1alpha1.CertConfig
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding certconfigs in namespace %#q", cr.Namespace))

		o := metav1.ListOptions{
			Continue:      "",
			LabelSelector: fmt.Sprintf("%s=%s", label.Cluster, key.ClusterID(&cr)),
			Limit:         listCertConfigLimit,
		}

		for {
			list, err := r.g8sClient.CoreV1alpha1().CertConfigs(cr.Namespace).List(o)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			for _, item := range list.Items {
				// Make a copy of an Item in order to not refer to loop iterator
				// variable. This is because we want to track a list of pointers.
				item := item
				certConfigs = append(certConfigs, &item)
			}

			o.Continue = list.Continue
			if o.Continue == "" {
				break
			}
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d certconfigs in namespace %#q", len(certConfigs), cr.Namespace))

	return certConfigs, nil
}
