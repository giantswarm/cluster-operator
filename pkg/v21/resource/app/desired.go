package app

import (
	"context"
	"strconv"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/annotation"
	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/pkg/v21/key"
	awskey "github.com/giantswarm/cluster-operator/service/controller/aws/v21/key"
	azurekey "github.com/giantswarm/cluster-operator/service/controller/azure/v21/key"
	kvmkey "github.com/giantswarm/cluster-operator/service/controller/kvm/v21/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) ([]*g8sv1alpha1.App, error) {
	clusterConfig, err := r.getClusterConfigFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var apps []*g8sv1alpha1.App

	for _, appSpec := range r.newAppSpecs() {
		apps = append(apps, r.newApp(clusterConfig, appSpec))
	}

	return apps, nil
}

func (r *Resource) newApp(clusterConfig v1alpha1.ClusterGuestConfig, appSpec key.AppSpec) *g8sv1alpha1.App {
	return &g8sv1alpha1.App{
		TypeMeta: metav1.TypeMeta{
			Kind:       "App",
			APIVersion: "application.giantswarm.io",
		},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				annotation.ForceHelmUpgrade: strconv.FormatBool(appSpec.UseUpgradeForce),
			},
			Labels: map[string]string{
				label.App:                appSpec.App,
				label.AppOperatorVersion: "1.0.0",
				label.Cluster:            clusterConfig.ID,
				label.ManagedBy:          project.Name(),
				label.Organization:       clusterConfig.Owner,
				label.ServiceType:        label.ServiceTypeManaged,
			},
			Name:      appSpec.App,
			Namespace: clusterConfig.ID,
		},
		Spec: g8sv1alpha1.AppSpec{
			Catalog:   appSpec.Catalog,
			Name:      appSpec.Chart,
			Namespace: appSpec.Namespace,
			Version:   appSpec.Version,

			Config: g8sv1alpha1.AppSpecConfig{
				ConfigMap: g8sv1alpha1.AppSpecConfigConfigMap{
					Name:      key.ClusterConfigMapName(clusterConfig),
					Namespace: clusterConfig.ID,
				},
			},

			KubeConfig: g8sv1alpha1.AppSpecKubeConfig{
				Context: g8sv1alpha1.AppSpecKubeConfigContext{
					Name: key.KubeConfigSecretName(clusterConfig),
				},
				InCluster: false,
				Secret: g8sv1alpha1.AppSpecKubeConfigSecret{
					Name:      key.KubeConfigSecretName(clusterConfig),
					Namespace: clusterConfig.ID,
				},
			},
		},
	}
}

func (r *Resource) newAppSpecs() []key.AppSpec {
	switch r.provider {
	case "aws":
		return append(key.CommonAppSpecs(), awskey.AppSpecs()...)
	case "azure":
		return append(key.CommonAppSpecs(), azurekey.AppSpecs()...)
	case "kvm":
		return append(key.CommonAppSpecs(), kvmkey.AppSpecs()...)
	default:
		return key.CommonAppSpecs()
	}
}
