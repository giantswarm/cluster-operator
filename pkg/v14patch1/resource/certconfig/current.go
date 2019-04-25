package certconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/v14/key"
)

// GetCurrentState takes observed custom object as an input and based on that
// information looks for current state of cluster certconfigs and returns them.
// Return value is of type []*v1alpha1.CertConfig.
func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	clusterGuestConfig, err := r.toClusterGuestConfigFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	objectMeta, err := r.toClusterObjectMetaFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "looking for a list of certconfigs in the Kubernetes API")

	var certConfigs []*v1alpha1.CertConfig
	{
		labelSelector := &metav1.LabelSelector{}
		labelSelector = metav1.AddLabelToSelector(labelSelector, label.LegacyClusterID, key.ClusterID(clusterGuestConfig))

		selector, err := metav1.LabelSelectorAsSelector(labelSelector)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		listOptions := metav1.ListOptions{
			LabelSelector: selector.String(),
			Limit:         listCertConfigLimit,
		}

		continueToken := ""

		for {
			listOptions.Continue = continueToken

			certConfigList, err := r.g8sClient.CoreV1alpha1().CertConfigs(objectMeta.Namespace).List(listOptions)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			for _, item := range certConfigList.Items {
				// Make a copy of an Item in order to not refer to loop
				// iterator variable.
				item := item
				certConfigs = append(certConfigs, &item)
			}

			continueToken = certConfigList.Continue
			if continueToken == "" {
				break
			}
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found a list of %d certconfigs in the Kubernetes API", len(certConfigs)))

	return certConfigs, nil
}
