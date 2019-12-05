package app

import (
	"context"
	"fmt"
	"strconv"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/cluster-operator/pkg/annotation"
	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	awskey "github.com/giantswarm/cluster-operator/service/controller/aws/key"
	azurekey "github.com/giantswarm/cluster-operator/service/controller/azure/key"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v22/key"
	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
	pkgkey "github.com/giantswarm/cluster-operator/service/controller/key"
	kvmkey "github.com/giantswarm/cluster-operator/service/controller/kvm/key"
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

	configMaps, err := r.getConfigMaps(ctx, cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	secrets, err := r.getSecrets(ctx, cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var apps []*g8sv1alpha1.App

	for _, appSpec := range r.newAppSpecs() {
		userConfig := newUserConfig(cr, appSpec, configMaps, secrets)

		if !appSpec.LegacyOnly {
			apps = append(apps, r.newApp(*cc, cr, appSpec, userConfig))
		}
	}

	return apps, nil
}

func (r *Resource) getConfigMaps(ctx context.Context, cr cmav1alpha1.Cluster) (map[string]corev1.ConfigMap, error) {
	configMaps := map[string]corev1.ConfigMap{}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding configMaps in namespace %#q", key.ClusterID(&cr)))

	list, err := r.k8sClient.CoreV1().ConfigMaps(key.ClusterID(&cr)).List(metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, cm := range list.Items {
		configMaps[cm.Name] = cm
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d configMaps in namespace %#q", len(configMaps), key.ClusterID(&cr)))

	return configMaps, nil
}

func (r *Resource) getSecrets(ctx context.Context, cr cmav1alpha1.Cluster) (map[string]corev1.Secret, error) {
	secrets := map[string]corev1.Secret{}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding secrets in namespace %#q", key.ClusterID(&cr)))

	list, err := r.k8sClient.CoreV1().Secrets(key.ClusterID(&cr)).List(metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, s := range list.Items {
		secrets[s.Name] = s
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d secrets in namespace %#q", len(secrets), key.ClusterID(&cr)))

	return secrets, nil
}

func (r *Resource) newApp(cc controllercontext.Context, cr cmav1alpha1.Cluster, appSpec pkgkey.AppSpec, userConfig g8sv1alpha1.AppSpecUserConfig) *g8sv1alpha1.App {
	configMapName := key.ClusterConfigMapName(&cr)

	// Override config map name when specified.
	if appSpec.ConfigMapName != "" {
		configMapName = appSpec.ConfigMapName
	}

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
					Name:      configMapName,
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

			UserConfig: userConfig,
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

func newUserConfig(cr cmav1alpha1.Cluster, appSpec pkgkey.AppSpec, configMaps map[string]corev1.ConfigMap, secrets map[string]corev1.Secret) g8sv1alpha1.AppSpecUserConfig {
	userConfig := g8sv1alpha1.AppSpecUserConfig{}

	_, ok := configMaps[pkgkey.AppUserConfigMapName(appSpec)]
	if ok {
		configMapSpec := g8sv1alpha1.AppSpecUserConfigConfigMap{
			Name:      pkgkey.AppUserConfigMapName(appSpec),
			Namespace: key.ClusterID(&cr),
		}

		userConfig.ConfigMap = configMapSpec
	}

	_, ok = secrets[pkgkey.AppUserSecretName(appSpec)]
	if ok {
		secretSpec := g8sv1alpha1.AppSpecUserConfigSecret{
			Name:      pkgkey.AppUserSecretName(appSpec),
			Namespace: key.ClusterID(&cr),
		}

		userConfig.Secret = secretSpec
	}

	return userConfig
}
