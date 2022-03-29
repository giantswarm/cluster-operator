package certconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/v6/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/v3/pkg/label"
	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
)

// getCurrentState takes observed custom object as an input and based on that
// information looks for current state of cluster CertConfig CRs and returns
// them. Return value is of type []*v1alpha1.CertConfig.
func (r *Resource) getCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var certConfigs []*v1alpha1.CertConfig
	{
		r.logger.Debugf(ctx, "finding CertConfig CRs in namespace %#q", cr.Namespace)

		o := metav1.ListOptions{
			Continue:      "",
			LabelSelector: fmt.Sprintf("%s=%s", label.Cluster, key.ClusterID(&cr)),
			Limit:         listCertConfigLimit,
		}

		for {
			list := &v1alpha1.CertConfigList{}
			err := r.ctrlClient.List(ctx, list, &client.ListOptions{Raw: &o})
			if err != nil {
				return nil, microerror.Mask(err)
			}

			for _, item := range list.Items {
				certConfigs = append(certConfigs, item.DeepCopy())
			}

			o.Continue = list.Continue
			if o.Continue == "" {
				break
			}
		}
	}

	r.logger.Debugf(ctx, "found %d CertConfig CRs in namespace %#q", len(certConfigs), cr.Namespace)

	return certConfigs, nil
}
