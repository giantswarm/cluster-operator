package app

import (
	"context"
	"strconv"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/cluster-operator/pkg/annotation"
	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	pkgkey "github.com/giantswarm/cluster-operator/pkg/v21/key"
	awskey "github.com/giantswarm/cluster-operator/service/controller/aws/v21/key"
	azurekey "github.com/giantswarm/cluster-operator/service/controller/azure/v21/key"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v21/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v21/key"
	kvmkey "github.com/giantswarm/cluster-operator/service/controller/kvm/v21/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) ([]*g8sv1alpha1.App, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var apps []*g8sv1alpha1.App

	for _, appSpec := range r.newAppSpecs() {
		apps = append(apps, r.newApp(*cc, cr, appSpec))
	}

	return apps, nil
}

func (r *Resource) newApp(cc controllercontext.Context, cr cmav1alpha1.Cluster, appSpec pkgkey.AppSpec) *g8sv1alpha1.App {
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
				label.AppOperatorVersion: cc.Status.Versions[label.AppOperatorVersion],
				label.Cluster:            key.ClusterID(&cr),
				label.ManagedBy:          project.Name(),
				label.Organization:       key.OrganizationID(&cr),
				label.ServiceType:        label.ServiceTypeManaged,
			},
			Name:      appSpec.App,
			Namespace: key.ClusterID(&cr),
		},
		Spec: g8sv1alpha1.AppSpec{
			Catalog:   appSpec.Catalog,
			Name:      appSpec.Chart,
			Namespace: appSpec.Namespace,
			Version:   appSpec.Version,

			Config: g8sv1alpha1.AppSpecConfig{
				ConfigMap: g8sv1alpha1.AppSpecConfigConfigMap{
					Name:      key.ClusterConfigMapName(&cr),
					Namespace: key.ClusterID(&cr),
				},
			},

			KubeConfig: g8sv1alpha1.AppSpecKubeConfig{
				Context: g8sv1alpha1.AppSpecKubeConfigContext{
					Name: key.KubeConfigSecretName(&cr),
				},
				Secret: g8sv1alpha1.AppSpecKubeConfigSecret{
					Name:      key.KubeConfigSecretName(&cr),
					Namespace: key.ClusterID(&cr),
				},
			},
		},
	}
}

func (r *Resource) newAppSpecs() []pkgkey.AppSpec {
	switch r.provider {
	case "aws":
		return append(pkgkey.CommonAppSpecs(), awskey.AppSpecs()...)
	case "azure":
		return append(pkgkey.CommonAppSpecs(), azurekey.AppSpecs()...)
	case "kvm":
		return append(pkgkey.CommonAppSpecs(), kvmkey.AppSpecs()...)
	default:
		return pkgkey.CommonAppSpecs()
	}
}
