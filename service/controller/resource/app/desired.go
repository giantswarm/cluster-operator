package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	g8sv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	apiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/v5/pkg/annotation"
	pkglabel "github.com/giantswarm/cluster-operator/v5/pkg/label"
	"github.com/giantswarm/cluster-operator/v5/pkg/project"
	"github.com/giantswarm/cluster-operator/v5/service/controller/key"
	"github.com/giantswarm/cluster-operator/v5/service/internal/releaseversion"
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

	appOperatorComponent := componentVersions[releaseversion.AppOperator]
	appOperatorVersion := appOperatorComponent.Version
	if appOperatorVersion == "" {
		return nil, microerror.Maskf(notFoundError, "%#q component version not found", releaseversion.AppOperator)
	}

	// Define app CR for app-operator in the management cluster namespace.
	appOperatorAppSpec := newAppOperatorAppSpec(cr, appOperatorComponent)
	apps = append(apps, r.newApp(key.UniqueOperatorVersion, cr, appOperatorAppSpec, g8sv1alpha1.AppSpecUserConfig{
		ConfigMap: g8sv1alpha1.AppSpecUserConfigConfigMap{
			Name:      "app-operator-konfigure",
			Namespace: "giantswarm",
		},
	}, nil))

	for _, appSpec := range appSpecs {
		userConfig := newUserConfig(cr, appSpec, configMaps, secrets)
		extraConfigs, err := r.getAppExtraConfigs(ctx, cr, appSpec)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		if !appSpec.LegacyOnly {
			apps = append(apps, r.newApp(appOperatorVersion, cr, appSpec, userConfig, extraConfigs))
		}
	}

	return apps, nil
}

func (r *Resource) getConfigMaps(ctx context.Context, cr apiv1beta1.Cluster) (map[string]corev1.ConfigMap, error) {
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

func (r *Resource) getSecrets(ctx context.Context, cr apiv1beta1.Cluster) (map[string]corev1.Secret, error) {
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

func (r *Resource) getUserOverrideConfig(ctx context.Context, cr apiv1beta1.Cluster) (userOverrideConfig, error) {
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

func (r *Resource) newApp(appOperatorVersion string, cr apiv1beta1.Cluster, appSpec key.AppSpec, userConfig g8sv1alpha1.AppSpecUserConfig, extraConfigs []g8sv1alpha1.AppExtraConfig) *g8sv1alpha1.App {
	configMapName := key.ClusterConfigMapName(&cr)

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
				Namespace: key.ClusterID(&cr),
			},
		}
	}

	var kubeConfig g8sv1alpha1.AppSpecKubeConfig

	if appSpec.InCluster || key.IsBundle(appName) {
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

	appNamespace := appSpec.Namespace
	// If the app is a bundle, we ensure the MC app operator deploys the apps
	// so the cluster-operator for the wc deploys the apps to the WC.
	desiredAppOperatorVersion := appOperatorVersion
	if key.IsBundle(appName) {
		appName = fmt.Sprintf("%s-%s", key.ClusterID(&cr), appName)
		desiredAppOperatorVersion = key.UniqueOperatorVersion
		appNamespace = key.ClusterID(&cr)
	}

	annotations := map[string]string{
		annotation.ForceHelmUpgrade: strconv.FormatBool(appSpec.UseUpgradeForce),
	}

	if len(appSpec.DependsOn) > 0 {
		annotations["app-operator.giantswarm.io/depends-on"] = strings.Join(appSpec.DependsOn, ",")
	}

	return &g8sv1alpha1.App{
		TypeMeta: metav1.TypeMeta{
			Kind:       "App",
			APIVersion: "application.giantswarm.io",
		},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: annotations,
			Labels: map[string]string{
				label.AppKubernetesName:  appSpec.App,
				label.AppOperatorVersion: desiredAppOperatorVersion,
				label.Cluster:            key.ClusterID(&cr),
				label.ManagedBy:          project.Name(),
				label.Organization:       key.OrganizationID(&cr),
				pkglabel.ServiceType:     pkglabel.ServiceTypeManaged,
			},
			Name:      appName,
			Namespace: key.ClusterID(&cr),
		},
		Spec: g8sv1alpha1.AppSpec{
			Catalog:      appSpec.Catalog,
			Name:         appSpec.Chart,
			Namespace:    appNamespace,
			Version:      appSpec.Version,
			Config:       config,
			ExtraConfigs: extraConfigs,
			KubeConfig:   kubeConfig,
			UserConfig:   userConfig,
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

	entries, ok := index.Entries[appNameWithAppSuffix]
	if ok && len(entries) > 0 {
		entry := findEntryByNameAndVersion(entries, appNameWithAppSuffix, version)
		if entry != nil {
			return entry.Name, nil
		}
	}

	entries, ok = index.Entries[appNameWithoutAppSuffix]
	if ok && len(entries) > 0 {
		entry := findEntryByNameAndVersion(entries, appNameWithoutAppSuffix, version)
		if entry != nil {
			return entry.Name, nil
		}
	}
	return "", microerror.Mask(fmt.Errorf("Could not find chart %s in %s catalog", appName, catalog))
}

func findEntryByNameAndVersion(entries []IndexEntry, appName string, appVersion string) *IndexEntry {
	for _, entry := range entries {
		if entry.Version == appVersion && entry.Name == appName {
			return &entry
		}
	}
	return nil
}

func (r *Resource) newAppSpecs(ctx context.Context, cr apiv1beta1.Cluster) ([]key.AppSpec, error) {
	userOverrideConfigs, err := r.getUserOverrideConfig(ctx, cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	apps, err := r.releaseVersion.Apps(ctx, &cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if r.provider == "aws" {
		awsCluster := &v1alpha3.AWSCluster{}
		err := r.ctrlClient.Get(ctx, types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, awsCluster)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		if key.IRSAEnabled(awsCluster) {
			// add IRSA app to the list
			version, err := getLatestVersion(ctx, r.ctrlClient, key.IRSAAppName, "default")
			if err != nil {
				return nil, microerror.Mask(err)
			}
			apps[key.IRSAAppName] = releaseversion.ReleaseApp{Catalog: key.IRSAAppCatalog, Version: version, DependsOn: []string{"cert-manager"}}
			r.logger.Debugf(ctx, "installing IRSA app")
		} else {
			r.logger.Debugf(ctx, "missing annotation for IRSA feature, not installing app")
		}
	} else {
		r.logger.Debugf(ctx, "not aws provider, skipping AWS IRSA check")
	}

	var specs []key.AppSpec
	for appName, app := range apps {
		var catalog string
		if app.Catalog == "" {
			catalog = r.defaultConfig.Catalog
		} else {
			catalog = app.Catalog
		}

		if !r.kiamWatchDogEnabled && appName == "kiam-watchdog" {
			// skip if disabled
			continue
		}

		chart, err := r.chartName(ctx, appName, catalog, app.Version)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		spec := key.AppSpec{
			App:             appName,
			Catalog:         catalog,
			Chart:           chart,
			DependsOn:       app.DependsOn,
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

func newAppOperatorAppSpec(cr apiv1beta1.Cluster, component releaseversion.ReleaseComponent) key.AppSpec {
	var operatorAppVersion string

	// Setting the reference allows us to deploy from a test catalog.
	if component.Reference != "" {
		operatorAppVersion = component.Reference
	} else {
		operatorAppVersion = component.Version
	}

	return key.AppSpec{
		App: releaseversion.AppOperator,
		// Override app name to include the cluster ID.
		AppName:         fmt.Sprintf("%s-%s", releaseversion.AppOperator, key.ClusterID(&cr)),
		Catalog:         component.Catalog,
		Chart:           releaseversion.AppOperator,
		InCluster:       true,
		Namespace:       key.ClusterID(&cr),
		UseUpgradeForce: false,
		Version:         operatorAppVersion,
	}
}

func newUserConfig(cr apiv1beta1.Cluster, appSpec key.AppSpec, configMaps map[string]corev1.ConfigMap, secrets map[string]corev1.Secret) g8sv1alpha1.AppSpecUserConfig {
	// User config naming is different for bundle apps.

	configMapName := key.AppUserConfigMapName(appSpec)
	{
		var appName string
		if appSpec.AppName != "" {
			appName = appSpec.AppName
		} else {
			appName = appSpec.App
		}
		if key.IsBundle(appName) {
			configMapName = fmt.Sprintf("%s-%s", key.ClusterID(&cr), configMapName)
		}
	}

	userConfig := g8sv1alpha1.AppSpecUserConfig{}
	_, ok := configMaps[configMapName]
	if ok {
		configMapSpec := g8sv1alpha1.AppSpecUserConfigConfigMap{
			Name:      configMapName,
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

func (r *Resource) getAppExtraConfigs(ctx context.Context, cr apiv1beta1.Cluster, appSpec key.AppSpec) ([]g8sv1alpha1.AppExtraConfig, error) {
	var err error
	var ret []g8sv1alpha1.AppExtraConfig

	r.logger.Debugf(ctx, "finding configMaps to be used as extraConfigs for app %q in namespace %#q", appSpec.App, key.ClusterID(&cr))

	configMaps := corev1.ConfigMapList{}
	err = r.ctrlClient.List(ctx, &configMaps, ctrlClient.MatchingLabels{label.AppKubernetesName: appSpec.App}, ctrlClient.InNamespace(key.ClusterID(&cr)))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, cm := range configMaps.Items {
		priority := g8sv1alpha1.ConfigPriorityDefault
		{
			priorityStr, found := cm.Annotations[annotation.AppConfigPriority]
			if found {
				priority, err = convertAndValidatePriority(priorityStr)
				if err != nil || priority <= 0 {
					r.logger.Debugf(ctx, "Invalid value for %q annotation in configMap %q. Should be a positive number. Defaulting to %d", annotation.AppConfigPriority, cm.Name, g8sv1alpha1.ConfigPriorityDefault)
					priority = g8sv1alpha1.ConfigPriorityDefault
				}
			}
		}

		r.logger.Debugf(ctx, "Using configMap %q as extraConfig with priority %d for app %q", cm.Name, priority, appSpec.App)
		ret = append(ret, g8sv1alpha1.AppExtraConfig{
			Kind:      "configMap",
			Name:      cm.Name,
			Namespace: cm.Namespace,
			Priority:  priority,
		})
	}

	r.logger.Debugf(ctx, "finding secrets to be used as extraConfigs for app %q in namespace %#q", appSpec.App, key.ClusterID(&cr))

	secrets := corev1.SecretList{}
	err = r.ctrlClient.List(ctx, &secrets, ctrlClient.MatchingLabels{label.AppKubernetesName: appSpec.App}, ctrlClient.InNamespace(key.ClusterID(&cr)))

	if err != nil {
		return nil, microerror.Mask(err)
	}
	for _, secret := range secrets.Items {
		priority := g8sv1alpha1.ConfigPriorityDefault
		{
			priorityStr, found := secret.Annotations[annotation.AppConfigPriority]
			if found {
				priority, err = convertAndValidatePriority(priorityStr)
				if err != nil || priority <= 0 {
					r.logger.Debugf(ctx, "Invalid value for %q annotation in secret %q. Should be a positive number. Defaulting to %d", annotation.AppConfigPriority, secret.Name, g8sv1alpha1.ConfigPriorityDefault)
					priority = g8sv1alpha1.ConfigPriorityDefault
				}
			}
		}

		r.logger.Debugf(ctx, "Using secret %q as extraConfig for app %q", secret.Name, appSpec.App)
		ret = append(ret, g8sv1alpha1.AppExtraConfig{
			Kind:      "secret",
			Name:      secret.Name,
			Namespace: secret.Namespace,
			Priority:  priority,
		})
	}

	return ret, nil
}

// See: https://docs.giantswarm.io/app-platform/app-configuration/#extra-configs
func convertAndValidatePriority(priorityStr string) (int, error) {
	priority, err := strconv.Atoi(priorityStr)

	if err != nil {
		return g8sv1alpha1.ConfigPriorityDefault, err
	}

	if priority > g8sv1alpha1.ConfigPriorityCatalog && priority <= g8sv1alpha1.ConfigPriorityMaximum {
		return priority, nil
	}

	return g8sv1alpha1.ConfigPriorityDefault, err
}

func (r *Resource) getCatalogIndex(ctx context.Context, catalogName string) ([]byte, error) {
	client := &http.Client{}

	var err error
	catalog := &g8sv1alpha1.AppCatalog{}
	{
		err = r.ctrlClient.Get(ctx, types.NamespacedName{Name: catalogName}, catalog)
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

		body, err = io.ReadAll(response.Body)
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

func getLatestVersion(ctx context.Context, ctrlClient ctrlClient.Client, app, catalog string) (string, error) {
	catalogEntryList := &g8sv1alpha1.AppCatalogEntryList{}
	err := ctrlClient.List(ctx, catalogEntryList, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{
			"app.kubernetes.io/name":            app,
			"application.giantswarm.io/catalog": catalog,
			"latest":                            "true",
		}), Namespace: "giantswarm"})
	if err != nil {
		return "", microerror.Mask(err)
	} else if len(catalogEntryList.Items) != 1 {
		// return default
		return key.IRSAAppVersion, nil
	}

	return catalogEntryList.Items[0].Spec.Version, nil
}
