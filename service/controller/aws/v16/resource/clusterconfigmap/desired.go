package clusterconfigmap

import (
	"context"

	"github.com/giantswarm/microerror"
	"k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/v16/key"
	awskey "github.com/giantswarm/cluster-operator/service/controller/aws/v16/key"
)

func (r *StateGetter) GetDesiredState(ctx context.Context, obj interface{}) ([]*v1.ConfigMap, error) {
	customObject, err := awskey.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterGuestConfig := awskey.ClusterGuestConfig(customObject)

	configMapName := key.ClusterConfigMapName(clusterGuestConfig)

	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: clusterGuestConfig.ID,
			Labels: map[string]string{
				label.Cluster:      clusterGuestConfig.ID,
				label.ManagedBy:    r.projectName,
				label.Organization: clusterGuestConfig.Owner,
				label.ServiceType:  label.ServiceTypeManaged,
			},
		},
		Data: map[string]string{
			"baseDomain": key.DNSZone(clusterGuestConfig),
		},
	}

	return []*corev1.ConfigMap{&cm}, nil
}
