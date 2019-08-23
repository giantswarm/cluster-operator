package configmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
)

func (s *Service) GetCurrentState(ctx context.Context, clusterConfig ClusterConfig) ([]*corev1.ConfigMap, error) {
	var currentConfigMaps []*corev1.ConfigMap

	// Namespaces used by all providers. Uses a map for deduping.
	namespaces := map[string]bool{
		metav1.NamespaceSystem: true,
	}

	// Add any provider specific namespaces.
	for _, namespace := range clusterConfig.Namespaces {
		namespaces[namespace] = true
	}

	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s, %s=%s", label.ServiceType, label.ServiceTypeManaged, label.ManagedBy, project.Name()),
	}

	for namespace := range namespaces {
		configMapList, err := cc.Client.TenantCluster.K8s.CoreV1().ConfigMaps(namespace).List(listOptions)
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
