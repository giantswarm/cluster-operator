package app

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/v18/key"
	awskey "github.com/giantswarm/cluster-operator/service/controller/aws/v18/key"
)

func (s *StateGetter) GetDesiredState(ctx context.Context, obj interface{}) ([]*v1alpha1.App, error) {
	customObject, err := awskey.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredApps := make([]*v1alpha1.App, 0)

	clusterGuestConfig := awskey.ClusterGuestConfig(customObject)
	clusterID := key.ClusterID(clusterGuestConfig)

	for _, spec := range awskey.AppSpecs() {
		app := &v1alpha1.App{
			ObjectMeta: metav1.ObjectMeta{
				Name:      spec.App,
				Namespace: clusterID,
				Labels: map[string]string{
					label.ManagedBy:   s.projectName,
					label.ServiceType: label.ServiceTypeManaged,
				},
			},
			Spec: v1alpha1.AppSpec{
				Catalog: spec.Chart,
				KubeConfig: v1alpha1.AppSpecKubeConfig{
					Context: v1alpha1.AppSpecKubeConfigContext{},
					Secret: v1alpha1.AppSpecKubeConfigSecret{
						Name:      key.KubeConfigSecretName(clusterGuestConfig),
						Namespace: clusterID,
					},
				},
				Name:      spec.App,
				Namespace: spec.Namespace,
				Version:   spec.Version,
			},
		}

		desiredApps = append(desiredApps, app)
	}

	return desiredApps, nil
}
