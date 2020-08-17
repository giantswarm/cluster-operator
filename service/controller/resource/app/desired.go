package app

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ghodss/yaml"
	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/clusterclient/service/release/searcher"
	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/annotation"
	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
	pkgerrors "github.com/giantswarm/cluster-operator/service/controller/internal/errors"
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

type appConfig struct {
	Catalog string `json:"catalog"`
	Version string `json:"version"`
}

type userOverrideConfig map[string]appConfig

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

	ch := make(chan struct{})

	var chartConfigList *v1alpha1.ChartConfigList

	go func() {
		chartConfigList, err = cc.Client.TenantCluster.G8s.CoreV1alpha1().ChartConfigs("giantswarm").List(listOptions)
		close(ch)
	}()

	select {
	case <-ch:
		// Fall through.
	case <-time.After(3 * time.Second):
		// Set status so we don't try to connect to the tenant cluster
		// again in this reconciliation loop.
		cc.Status.TenantCluster.IsUnavailable = true

		r.logger.LogCtx(ctx, "level", "debug", "message", "timeout getting chartconfig crs")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil, nil
	}

	if tenant.IsAPINotAvailable(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is not available")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil
	} else if pkgerrors.IsChartConfigNotAvailable(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "chartconfig CRs are not available")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil
	} else if pkgerrors.IsChartConfigNotInstalled(err) {
		// chartconfig CRD is not installed. So this cluster does not need
		// any CRs to be migrated.
		r.logger.LogCtx(ctx, "level", "debug", "message", "chartconfig CRD not installed")
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	chartConfigs := map[string]v1alpha1.ChartConfig{}

	if err == nil {
		for _, cr := range chartConfigList.Items {
			// Chartconfig and app CRs have the same app label. So we use this as
			// the key for the map.
			appName := cr.Labels[label.App]
			chartConfigs[appName] = cr
		}
	}

	// Get all configmaps in kube-system in the tenant cluster to ensure user
	// configmaps have been migrated.
	configMapList, err := cc.Client.TenantCluster.K8s.CoreV1().ConfigMaps(metav1.NamespaceSystem).List(listOptions)
	if tenant.IsAPINotAvailable(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is not available")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	tenantConfigMaps := map[string]corev1.ConfigMap{}

	for _, cm := range configMapList.Items {
		tenantConfigMaps[cm.Name] = cm
	}

	var apps []*g8sv1alpha1.App
	appSpecs, err := r.newAppSpecs(ctx, clusterConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	appOperatorVersion, err := r.getComponentVersion(clusterConfig.ReleaseVersion, "app-operator")
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, appSpec := range appSpecs {
		if !hasMigrationCompleted(appSpec, chartConfigs, configMaps, tenantConfigMaps) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("app %#q user values not migrated yet, continuing", appSpec.App))
			continue
		}

		userConfig, err := newUserConfig(clusterConfig, appSpec, configMaps, secrets)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		if !appSpec.ClusterAPIOnly {
			app := r.newApp(clusterConfig, appSpec, userConfig, appOperatorVersion)

			apps = append(apps, app)
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

func (r *Resource) getUserOverrideConfig(ctx context.Context, clusterConfig v1alpha1.ClusterGuestConfig) (userOverrideConfig, error) {
	userConfig, err := r.k8sClient.CoreV1().ConfigMaps(key.ClusterID(clusterConfig)).Get("user-override-apps", metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		// fall through
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	u := userOverrideConfig{}

	appConfigs, ok := userConfig.Data[clusterConfig.ReleaseVersion]
	if !ok {
		// no release override configs, fall through
		return nil, nil
	}

	err = yaml.Unmarshal([]byte(appConfigs), &u)
	if err != nil {
		r.logger.LogCtx(ctx, "level", "error", "message", "failed to unmarshal the user config", "stack", microerror.Stack(err))
		return nil, nil
	}

	return u, nil
}

func (r *Resource) newApp(clusterConfig v1alpha1.ClusterGuestConfig, appSpec key.AppSpec, userConfig g8sv1alpha1.AppSpecUserConfig, appOperatorVersion string) *g8sv1alpha1.App {
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
				label.AppOperatorVersion: appOperatorVersion,
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

	userOverrideConfigs, err := r.getUserOverrideConfig(ctx, cr)
	if err != nil {
		return nil, microerror.Mask(err)
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

		// To test apps in the testing catalog, users can override default app properties with
		// a user-override-apps configmap.
		if val, ok := userOverrideConfigs[app.App]; ok {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found a user override app config for %#q, applying it", app.App))
			if val.Catalog != "" {
				spec.Catalog = val.Catalog
			}
			if val.Version != "" {
				spec.Version = val.Version
			}
		}

		specs = append(specs, spec)
	}
	return specs, nil
}

func (r *Resource) getComponentVersion(releaseVersion, component string) (string, error) {
	release, err := r.g8sClient.ReleaseV1alpha1().Releases().Get(releaseVersion, metav1.GetOptions{})
	if err != nil {
		return "", microerror.Mask(err)
	}

	for _, c := range release.Spec.Components {
		if c.Name == component {
			return c.Version, nil
		}
	}

	return "", microerror.Maskf(notFoundError, fmt.Sprintf("can't find the release version %#q", releaseVersion))
}

// hasMigrationCompleted checks if the migration from chartconfig to app CR
// has completed. We delay creating the app CR until any user settings have
// been copied to their new location.
func hasMigrationCompleted(appSpec key.AppSpec, chartConfigs map[string]v1alpha1.ChartConfig, configMaps, tenantConfigMaps map[string]corev1.ConfigMap) bool {
	_, hasChartConfig := chartConfigs[appSpec.App]

	// No chartconfig CR exists so either no migration was needed or it has
	// completed.
	if !hasChartConfig {
		return true
	}

	tenantCM, tenantExists := tenantConfigMaps[key.AppUserConfigMapName(appSpec)]
	_, cmExists := configMaps[key.AppUserConfigMapName(appSpec)]

	// A tenant configmap exists with user settings but the user configmap for
	// this app CR does not exist yet. We delay creating the app CR until it
	// does so the app is installed with the correct settings.
	if tenantExists && len(tenantCM.Data) > 0 && !cmExists {
		return false
	}

	return true
}

func newUserConfig(clusterConfig v1alpha1.ClusterGuestConfig, appSpec key.AppSpec, configMaps map[string]corev1.ConfigMap, secrets map[string]corev1.Secret) (g8sv1alpha1.AppSpecUserConfig, error) {
	userConfig := g8sv1alpha1.AppSpecUserConfig{}

	_, cmExists := configMaps[key.AppUserConfigMapName(appSpec)]
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
