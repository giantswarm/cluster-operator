package app

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/annotation"
	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/pkg/v21/controllercontext"
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

	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", label.ManagedBy, project.Name()),
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Get all configmaps in kube-system in the tenant cluster to ensure user
	// configmaps have been migrated.
	list, err := cc.Client.TenantCluster.K8s.CoreV1().ConfigMaps(metav1.NamespaceSystem).List(listOptions)
	if tenant.IsAPINotAvailable(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is not available yet")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	} else if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "timeout getting chartconfig CRs")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil
	}

	tenantConfigMaps := map[string]corev1.ConfigMap{}

	for _, cm := range list.Items {
		configMaps[cm.Name] = cm
	}

	var apps []*g8sv1alpha1.App

	for _, appSpec := range r.newAppSpecs() {
		userConfig, err := newUserConfig(clusterConfig, appSpec, configMaps, tenantConfigMaps, secrets)
		if isNotMigratedError(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("app %#q not migrated yet, continuing", appSpec.App))
			continue
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		apps = append(apps, r.newApp(clusterConfig, appSpec, userConfig))
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

			UserConfig: userConfig,
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

func newUserConfig(clusterConfig v1alpha1.ClusterGuestConfig, appSpec key.AppSpec, configMaps, tenantConfigMaps map[string]corev1.ConfigMap, secrets map[string]corev1.Secret) (g8sv1alpha1.AppSpecUserConfig, error) {
	userConfig := g8sv1alpha1.AppSpecUserConfig{}

	tenantCM, tenantExists := tenantConfigMaps[key.AppUserConfigMapName(appSpec)]
	_, cmExists := configMaps[key.AppUserConfigMapName(appSpec)]

	// A tenant configmap exists with user settings but the user configmap for
	// this app CR does not exist yet. We delay creating the app CR until it
	// does so the app is installed with the correct values.
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
