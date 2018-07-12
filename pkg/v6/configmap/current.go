package configmap

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Service) GetCurrentState(ctx context.Context, config ConfigMapConfig) ([]*corev1.ConfigMap, error) {
	var currentConfigMaps []*corev1.ConfigMap

	guestK8sClient, err := s.guest.NewK8sClient(ctx, config.ClusterID, config.GuestAPIDomain)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Namespaces used by all providers. Uses a map for deduping.
	namespaces := map[string]bool{
		metav1.NamespaceSystem: true,
	}

	// Add any provider specific namespaces.
	for _, namespace := range config.Namespaces {
		namespaces[namespace] = true
	}

	for namespace := range namespaces {
		configMapList, err := guestK8sClient.CoreV1().ConfigMaps(namespace).List(metav1.ListOptions{})
		if err != nil {
			return nil, microerror.Mask(err)
		}

		for _, item := range configMapList.Items {
			c := item.DeepCopy()
			currentConfigMaps = append(currentConfigMaps, c)
		}
	}

	return currentConfigMaps, nil
}
