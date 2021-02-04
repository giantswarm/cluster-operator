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
	g8sv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/cluster-operator/v3/pkg/annotation"
	pkglabel "github.com/giantswarm/cluster-operator/v3/pkg/label"
	"github.com/giantswarm/cluster-operator/v3/pkg/project"
	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
	"github.com/giantswarm/cluster-operator/v3/service/internal/releaseversion"
)

type appConfig struct {
	Catalog string `json:"catalog"`
	Version string `json:"version"`
}

type userOverrideConfig map[string]appConfig

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) ([]*g8sv1alpha1.App, error) {
	cr, err := key.ToCluster(obj)
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

	componentVersions, err := r.releaseVersion.ComponentVersion(ctx, &cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	appOperatorVersion := componentVersions[releaseversion.AppOperator]

	for _, appSpec := range appSpecs {
		userConfig := newUserConfig(cr, appSpec, configMaps, secrets)

		if !appSpec.LegacyOnly {
			apps = append(apps, r.newApp(appOperatorVersion, cr, appSpec, userConfig))
		}
	}

	appOperatorSpec := key.AppSpec{
		App:             fmt.Sprintf("app-operator-%s", key.ClusterID(&cr)),
		Catalog:         "control-plane-test-catalog",
		Chart:           "app-operator",
		InCluster:       true,
		Namespace:       key.ClusterID(&cr),
		UseUpgradeForce: true,
		Version:         "3.1.0-77d0f102d4f1773cc48b03b4397ce1c7d003a090",
	}
	apps = append(apps, r.newApp("0.0.0", cr, appOperatorSpec, g8sv1alpha1.AppSpecUserConfig{}))

	return apps, nil
}

func (r *Resource) getConfigMaps(ctx context.Context, cr apiv1alpha2.Cluster) (map[string]corev1.ConfigMap, error) {
	configMaps := map[string]corev1.ConfigMap{}

	r.logger.Debugf(ctx, "finding configMaps in namespace %#q", key.ClusterID(&cr))

	list, err := r.k8sClient.CoreV1().ConfigMaps(key.ClusterID(&cr)).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, cm := range list.Items {
		configMaps[cm.Name] = cm
	}

	r.logger.Debugf(ctx, "found %d configMaps in namespace %#q", len(configMaps), key.ClusterID(&cr))

	return configMaps, nil
}

func (r *Resource) getSecrets(ctx context.Context, cr apiv1alpha2.Cluster) (map[string]corev1.Secret, error) {
	secrets := map[string]corev1.Secret{}

	r.logger.Debugf(ctx, "finding secrets in namespace %#q", key.ClusterID(&cr))

	list, err := r.k8sClient.CoreV1().Secrets(key.ClusterID(&cr)).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, s := range list.Items {
		secrets[s.Name] = s
	}

	r.logger.Debugf(ctx, "found %d secrets in namespace %#q", len(secrets), key.ClusterID(&cr))

	return secrets, nil
}

func (r *Resource) getUserOverrideConfig(ctx context.Context, cr apiv1alpha2.Cluster) (userOverrideConfig, error) {
	userConfig, err := r.k8sClient.CoreV1().ConfigMaps(key.ClusterID(&cr)).Get(ctx, "user-override-apps", metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		// fall through
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	u := userOverrideConfig{}

	appConfigs, ok := userConfig.Data[key.ReleaseVersion(&cr)]
	if !ok {
		// no release override configs, fall through
		return nil, nil
	}

	err = yaml.Unmarshal([]byte(appConfigs), &u)
	if err != nil {
		r.logger.Errorf(ctx, err, "failed to unmarshal the user config")
		return nil, nil
	}

	return u, nil
}

func (r *Resource) newApp(appOperatorVersion string, cr apiv1alpha2.Cluster, appSpec key.AppSpec, userConfig g8sv1alpha1.AppSpecUserConfig) *g8sv1alpha1.App {
	configMapName := key.ClusterConfigMapName(&cr)

	// Override config map name when specified.
	if appSpec.ConfigMapName != "" {
		configMapName = appSpec.ConfigMapName
	}

	var kubeConfig g8sv1alpha1.AppSpecKubeConfig

	if appSpec.InCluster {
		kubeConfig = g8sv1alpha1.AppSpecKubeConfig{
			InCluster: true,
		}
	} else {
		kubeConfig = g8sv1alpha1.AppSpecKubeConfig{
			Context: g8sv1alpha1.AppSpecKubeConfigContext{
				Name: key.KubeConfigSecretName(&cr),
			},
			Secret: g8sv1alpha1.AppSpecKubeConfigSecret{
				Name:      key.KubeConfigSecretName(&cr),
				Namespace: key.ClusterID(&cr),
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
				label.Cluster:            key.ClusterID(&cr),
				label.ManagedBy:          project.Name(),
				label.Organization:       key.OrganizationID(&cr),
				pkglabel.ServiceType:     pkglabel.ServiceTypeManaged,
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

func (r *Resource) newAppSpecs(ctx context.Context, cr apiv1alpha2.Cluster) ([]key.AppSpec, error) {
	userOverrideConfigs, err := r.getUserOverrideConfig(ctx, cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	apps, err := r.releaseVersion.Apps(ctx, &cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var specs []key.AppSpec
	for appName, app := range apps {
		var catalog string
		if app.Catalog == "" {
			catalog = r.defaultConfig.Catalog
		} else {
			catalog = app.Catalog
		}

		chart, err := r.chartName(ctx, appName, catalog, app.Version)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		spec := key.AppSpec{
			App:             appName,
			Catalog:         catalog,
			Chart:           chart,
			Namespace:       r.defaultConfig.Namespace,
			UseUpgradeForce: r.defaultConfig.UseUpgradeForce,
			Version:         app.Version,
		}
		// For some apps we can't use default settings. We check ConfigExceptions map
		// for these differences.
		// We are looking into ConfigException map to see if this chart is the case.
		if val, ok := r.overrideConfig[appName]; ok {
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

		// To test apps in the testing catalog, users can override default app properties with
		// a user-override-apps configmap.
		if val, ok := userOverrideConfigs[appName]; ok {
			r.logger.Debugf(ctx, "found a user override app config for %#q, applying it", appName)
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

func (r *Resource) getCatalogIndex(ctx context.Context, catalogName string) ([]byte, error) {
	client := &http.Client{}

	var err error
	var catalog *g8sv1alpha1.AppCatalog
	{
		catalog, err = r.g8sClient.ApplicationV1alpha1().AppCatalogs().Get(ctx, catalogName, metav1.GetOptions{})
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
