package app

import (
	"context"
	"fmt"
	"strconv"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/cluster-operator/pkg/annotation"
	pkgapp "github.com/giantswarm/cluster-operator/pkg/app"
	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/key"
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

	appSpecs, err := r.newAppSpecs(ctx, cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, appSpec := range appSpecs {
		userConfig := newUserConfig(cr, appSpec, configMaps, secrets)

		if !appSpec.LegacyOnly {
			apps = append(apps, r.newApp(*cc, cr, appSpec, userConfig))
		}
	}

	return apps, nil
}

func (r *Resource) getConfigMaps(ctx context.Context, cr apiv1alpha2.Cluster) (map[string]corev1.ConfigMap, error) {
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

func (r *Resource) getSecrets(ctx context.Context, cr apiv1alpha2.Cluster) (map[string]corev1.Secret, error) {
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

func (r *Resource) newApp(cc controllercontext.Context, cr apiv1alpha2.Cluster, appSpec key.AppSpec, userConfig g8sv1alpha1.AppSpecUserConfig) *g8sv1alpha1.App {
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

func (r *Resource) newAppSpecs(ctx context.Context, cr apiv1alpha2.Cluster) ([]key.AppSpec, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var specs []key.AppSpec

	for _, app := range cc.Status.Apps {
		spec := key.AppSpec{
			App:             app.App,
			Catalog:         pkgapp.Default.Catalog,
			Chart:           fmt.Sprintf("%s-app", app.App),
			Namespace:       pkgapp.Default.Namespace,
			UseUpgradeForce: pkgapp.Default.UseUpgradeForce,
			Version:         app.Version,
		}
		// Some apps may need to apply other chart name format, namespace, or helm properties.
		// We are looking into ConfigException map to see if this chart is the case.
		if val, ok := pkgapp.ConfigExceptions[app.App]; ok {
			if val.Chart != "" {
				spec.Chart = val.Chart
			}
			if val.Namespace != "" {
				spec.Namespace = val.Namespace
			}
			if !val.UseUpgradeForce {
				spec.UseUpgradeForce = false
			}
		}

		specs = append(specs, spec)
	}
	return specs, nil
}

func newUserConfig(cr apiv1alpha2.Cluster, appSpec key.AppSpec, configMaps map[string]corev1.ConfigMap, secrets map[string]corev1.Secret) g8sv1alpha1.AppSpecUserConfig {
	userConfig := g8sv1alpha1.AppSpecUserConfig{}

	_, ok := configMaps[key.AppUserConfigMapName(appSpec)]
	if ok {
		configMapSpec := g8sv1alpha1.AppSpecUserConfigConfigMap{
			Name:      key.AppUserConfigMapName(appSpec),
			Namespace: key.ClusterID(&cr),
		}

		userConfig.ConfigMap = configMapSpec
	}

	_, ok = secrets[key.AppUserSecretName(appSpec)]
	if ok {
		secretSpec := g8sv1alpha1.AppSpecUserConfigSecret{
			Name:      key.AppUserSecretName(appSpec),
			Namespace: key.ClusterID(&cr),
		}

		userConfig.Secret = secretSpec
	}

	return userConfig
}
