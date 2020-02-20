package app

import (
	"context"
	"fmt"
	"strconv"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/clusterclient/service/release/searcher"
	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/annotation"
	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) ([]*g8sv1alpha1.App, error) {
	clusterConfig, err := r.getClusterConfigFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	configMaps, err := r.getConfigMaps(ctx, clusterConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	secrets, err := r.getSecrets(ctx, clusterConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// TODO: Remove connection to tenant cluster once all tenant clusters use
	// app CRs instead of chartconfig CRs.
	//
	//	https://github.com/giantswarm/giantswarm/issues/7402
	//
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if cc.Client.TenantCluster.G8s == nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant clients not available")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil
	}

	if cc.Status.TenantCluster.IsUnavailable {
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is unavailable")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil
	}

	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", label.ManagedBy, project.Name()),
	}

	// Get all configmaps in kube-system in the tenant cluster to ensure user
	// configmaps have been migrated.
	list, err := cc.Client.TenantCluster.K8s.CoreV1().ConfigMaps(metav1.NamespaceSystem).List(listOptions)
	if tenant.IsAPINotAvailable(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is not available")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	tenantConfigMaps := map[string]corev1.ConfigMap{}

	for _, cm := range list.Items {
		tenantConfigMaps[cm.Name] = cm
	}

	var apps []*g8sv1alpha1.App
	appSpecs, err := r.newAppSpecs(ctx, clusterConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, appSpec := range appSpecs {
		userConfig, err := newUserConfig(clusterConfig, appSpec, configMaps, tenantConfigMaps, secrets)
		if IsNotMigratedError(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("app %#q user values not migrated yet, continuing", appSpec.App))
			continue
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		if !appSpec.ClusterAPIOnly {
			apps = append(apps, r.newApp(clusterConfig, appSpec, userConfig))
		}
	}

	return apps, nil
}

func (r *Resource) getConfigMaps(ctx context.Context, clusterConfig v1alpha1.ClusterGuestConfig) (map[string]corev1.ConfigMap, error) {
	configMaps := map[string]corev1.ConfigMap{}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding configMaps in namespace %#q", clusterConfig.ID))

	list, err := r.k8sClient.CoreV1().ConfigMaps(clusterConfig.ID).List(metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, cm := range list.Items {
		configMaps[cm.Name] = cm
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d configMaps in namespace %#q", len(configMaps), clusterConfig.ID))

	return configMaps, nil
}

func (r *Resource) getSecrets(ctx context.Context, clusterConfig v1alpha1.ClusterGuestConfig) (map[string]corev1.Secret, error) {
	secrets := map[string]corev1.Secret{}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding secrets in namespace %#q", clusterConfig.ID))

	list, err := r.k8sClient.CoreV1().Secrets(clusterConfig.ID).List(metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, s := range list.Items {
		secrets[s.Name] = s
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d secrets in namespace %#q", len(secrets), clusterConfig.ID))

	return secrets, nil
}

func (r *Resource) newApp(clusterConfig v1alpha1.ClusterGuestConfig, appSpec key.AppSpec, userConfig g8sv1alpha1.AppSpecUserConfig) *g8sv1alpha1.App {
	configMapName := key.ClusterConfigMapName(clusterConfig)

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
					Name:      configMapName,
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

			UserConfig: userConfig,
		},
	}
}

func (r *Resource) newAppSpecs(ctx context.Context, cr v1alpha1.ClusterGuestConfig) ([]key.AppSpec, error) {
	req := searcher.Request{
		ReleaseVersion: cr.ReleaseVersion,
	}

	res, err := r.clusterClient.Release.Searcher.Search(ctx, req)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	if len(res.Apps) == 0 {
		return nil, microerror.Maskf(executionFailedError, "no apps in release %#q", req.ReleaseVersion)
	}

	var specs []key.AppSpec

	for _, app := range res.Apps {
		spec := key.AppSpec{
			App:             app.App,
			Catalog:         r.defaultConfig.Catalog,
			Chart:           fmt.Sprintf("%s-app", app.App),
			Namespace:       r.defaultConfig.Namespace,
			UseUpgradeForce: r.defaultConfig.UseUpgradeForce,
			Version:         app.Version,
		}
		// For some apps we can't use default settings. We check ConfigExceptions map
		// for these differences.
		// We are looking into ConfigException map to see if this chart is the case.
		if val, ok := r.overrideConfig[app.App]; ok {
			if val.Chart != "" {
				spec.Chart = val.Chart
			}
			if val.Namespace != "" {
				spec.Namespace = val.Namespace
			}
			if val.UseUpgradeForce != nil {
				spec.UseUpgradeForce = *val.UseUpgradeForce
			}
		}

		// Nginx Ingress Controller uses its own configmap that includes the number of workers.
		if app.App == "nginx-ingress-controller" {
			spec.ConfigMapName = key.IngressControllerConfigMapName
		}

		specs = append(specs, spec)
	}
	return specs, nil
}

func newUserConfig(clusterConfig v1alpha1.ClusterGuestConfig, appSpec key.AppSpec, configMaps, tenantConfigMaps map[string]corev1.ConfigMap, secrets map[string]corev1.Secret) (g8sv1alpha1.AppSpecUserConfig, error) {
	userConfig := g8sv1alpha1.AppSpecUserConfig{}

	_, cmExists := configMaps[key.AppUserConfigMapName(appSpec)]
	tenantCM, tenantExists := tenantConfigMaps[key.AppUserConfigMapName(appSpec)]

	// A tenant configmap exists with user settings but the user configmap for
	// this app CR does not exist yet. We delay creating the app CR until it
	// does so the app is installed with the correct settings.
	if tenantExists && len(tenantCM.Data) > 0 && !cmExists {
		return g8sv1alpha1.AppSpecUserConfig{}, microerror.Maskf(notMigratedError, "%#q not migrated yet", appSpec.App)
	}

	if cmExists {
		configMapSpec := g8sv1alpha1.AppSpecUserConfigConfigMap{
			Name:      key.AppUserConfigMapName(appSpec),
			Namespace: clusterConfig.ID,
		}

		userConfig.ConfigMap = configMapSpec
	}

	_, secretExists := secrets[key.AppUserSecretName(appSpec)]
	if secretExists {
		secretSpec := g8sv1alpha1.AppSpecUserConfigSecret{
			Name:      key.AppUserSecretName(appSpec),
			Namespace: clusterConfig.ID,
		}

		userConfig.Secret = secretSpec
	}

	return userConfig, nil
}
