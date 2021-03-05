package app

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/clusterclient/service/release/searcher"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/annotation"
	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
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

	var apps []*g8sv1alpha1.App
	appSpecs, err := r.newAppSpecs(ctx, clusterConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// put `v` as a prefix of release version since all releases CRs keep this format.
	appOperatorComponent, err := r.getReleaseComponent(fmt.Sprintf("v%s", clusterConfig.ReleaseVersion), appOperatorComponentName)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	appOperatorVersion := appOperatorComponent.Version
	if appOperatorVersion == "" {
		return nil, microerror.Maskf(notFoundError, "%#q release component not found", appOperatorComponentName)
	}

	// Define app CR for app-operator in the management cluster namespace.
	appOperatorAppSpec := newAppOperatorAppSpec(clusterConfig, appOperatorComponent)
	apps = append(apps, r.newApp(clusterConfig, appOperatorAppSpec, g8sv1alpha1.AppSpecUserConfig{}, uniqueOperatorVersion))

	for _, appSpec := range appSpecs {
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
		r.logger.LogCtx(ctx, "level", "error", "message", "failed to unmarshal the user config", "stack", microerror.JSON(err))
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

	var appName string

	if appSpec.AppName != "" {
		appName = appSpec.AppName
	} else {
		appName = appSpec.App
	}

	var config g8sv1alpha1.AppSpecConfig

	if appSpec.InCluster {
		config = g8sv1alpha1.AppSpecConfig{}
	} else {
		config = g8sv1alpha1.AppSpecConfig{
			ConfigMap: g8sv1alpha1.AppSpecConfigConfigMap{
				Name:      configMapName,
				Namespace: clusterConfig.ID,
			},
		}
	}

	var kubeConfig g8sv1alpha1.AppSpecKubeConfig

	if appSpec.InCluster {
		kubeConfig = g8sv1alpha1.AppSpecKubeConfig{
			InCluster: true,
		}
	} else {
		kubeConfig = g8sv1alpha1.AppSpecKubeConfig{
			Context: g8sv1alpha1.AppSpecKubeConfigContext{
				Name: key.KubeConfigSecretName(clusterConfig),
			},
			Secret: g8sv1alpha1.AppSpecKubeConfigSecret{
				Name:      key.KubeConfigSecretName(clusterConfig),
				Namespace: clusterConfig.ID,
			},
		}
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
				label.AppKubernetesName:  appSpec.App,
				label.AppOperatorVersion: appOperatorVersion,
				label.Cluster:            clusterConfig.ID,
				label.ManagedBy:          project.Name(),
				label.Organization:       clusterConfig.Owner,
				label.ServiceType:        label.ServiceTypeManaged,
			},
			Name:      appName,
			Namespace: clusterConfig.ID,
		},
		Spec: g8sv1alpha1.AppSpec{
			Catalog:    appSpec.Catalog,
			Name:       appSpec.Chart,
			Namespace:  appSpec.Namespace,
			Version:    appSpec.Version,
			Config:     config,
			KubeConfig: kubeConfig,
			UserConfig: userConfig,
		},
	}
}

func (r *Resource) chartName(ctx context.Context, appName, catalog, version string) (string, error) {
	var index Index
	{
		indexYamlBytes, err := r.getCatalogIndex(ctx, catalog)
		if err != nil {
			return "", microerror.Mask(err)
		}

		err = yaml.Unmarshal(indexYamlBytes, &index)
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	appNameWithoutAppSuffix := strings.TrimSuffix(appName, "-app")
	appNameWithAppSuffix := fmt.Sprintf("%s-app", appNameWithoutAppSuffix)
	chartName := ""

	entries, ok := index.Entries[appNameWithAppSuffix]
	if !ok || len(entries) == 0 {
		entries, ok = index.Entries[appNameWithoutAppSuffix]
		if !ok || len(entries) == 0 {
			return "", microerror.Mask(fmt.Errorf("Could not find chart %s in %s catalog", appName, catalog))
		}
		chartName = appNameWithoutAppSuffix
	} else {
		chartName = appNameWithAppSuffix
	}

	for _, entry := range entries {
		if entry.Version == version && entry.Name == chartName {
			return entry.Name, nil
		}
	}

	return "", microerror.Mask(fmt.Errorf("Could not find chart %s in %s catalog", appName, catalog))
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
		chart, err := r.chartName(ctx, app.App, r.defaultConfig.Catalog, app.Version)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		spec := key.AppSpec{
			App:             app.App,
			Catalog:         r.defaultConfig.Catalog,
			Chart:           chart,
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

func newAppOperatorAppSpec(clusterConfig v1alpha1.ClusterGuestConfig, component releasev1alpha1.ReleaseSpecComponent) key.AppSpec {
	return key.AppSpec{
		App: appOperatorComponentName,
		// Override app name to include the cluster ID.
		AppName:         fmt.Sprintf("%s-%s", appOperatorComponentName, clusterConfig.ID),
		Catalog:         controlPlaneCatalog,
		Chart:           appOperatorComponentName,
		InCluster:       true,
		Namespace:       clusterConfig.ID,
		UseUpgradeForce: false,
		Version:         component.Version,
	}
}

func (r *Resource) getReleaseComponent(releaseVersion, component string) (releasev1alpha1.ReleaseSpecComponent, error) {
	release, err := r.g8sClient.ReleaseV1alpha1().Releases().Get(releaseVersion, metav1.GetOptions{})
	if err != nil {
		return releasev1alpha1.ReleaseSpecComponent{}, microerror.Mask(err)
	}

	for _, c := range release.Spec.Components {
		if c.Name == component {
			return c, nil
		}
	}

	return releasev1alpha1.ReleaseSpecComponent{}, microerror.Maskf(notFoundError, fmt.Sprintf("can't find the %#q component for %#q", component, releaseVersion))
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

func (r *Resource) getCatalogIndex(ctx context.Context, catalogName string) ([]byte, error) {
	client := &http.Client{}

	var err error
	var catalog *g8sv1alpha1.AppCatalog
	{
		catalog, err = r.g8sClient.ApplicationV1alpha1().AppCatalogs().Get(catalogName, metav1.GetOptions{})
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	url := strings.TrimRight(catalog.Spec.Storage.URL, "/") + "/index.yaml"
	body := []byte{}

	o := func() error {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, &bytes.Buffer{}) // nolint: gosec
		if err != nil {
			return microerror.Mask(err)
		}
		response, err := client.Do(request)
		if err != nil {
			return microerror.Mask(err)
		}
		defer response.Body.Close()

		body, err = ioutil.ReadAll(response.Body)
		if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}
	b := backoff.NewExponential(30*time.Second, 5*time.Second)
	n := backoff.NewNotifier(r.logger, ctx)

	err = backoff.RetryNotify(o, b, n)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return body, nil
}
