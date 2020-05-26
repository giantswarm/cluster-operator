package configmapmigration

import (
	"context"
	"fmt"
	"strings"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	awskey "github.com/giantswarm/cluster-operator/service/controller/aws/key"
	azurekey "github.com/giantswarm/cluster-operator/service/controller/azure/key"
	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
	pkgerrors "github.com/giantswarm/cluster-operator/service/controller/internal/errors"
	"github.com/giantswarm/cluster-operator/service/controller/key"
	kvmkey "github.com/giantswarm/cluster-operator/service/controller/kvm/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	objectMeta, err := r.getClusterObjectMetaFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// Cluster is being deleted. No migration is necessary.
	if key.IsDeleted(objectMeta) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "cluster is being deleted")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	if cc.Status.TenantCluster.IsUnavailable {
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is unavailable")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	chartSpecsToMigrate := r.newChartSpecsToMigrate()

	if len(chartSpecsToMigrate) == 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no charts to migrate")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	cr, err := r.getClusterConfigFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// Get all configmaps in the cluster namespace.
	clusterConfigMaps, err := r.k8sClient.CoreV1().ConfigMaps(key.ClusterID(cr)).List(metav1.ListOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	if cc.Client.TenantCluster.G8s == nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant clients not available")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", label.ManagedBy, project.Name()),
	}

	// Get all chartconfig CRs in the tenant cluster. The migration needs to
	// complete before we create app CRs. So we cancel the entire loop on error.
	chartConfigs, err := cc.Client.TenantCluster.G8s.CoreV1alpha1().ChartConfigs("giantswarm").List(listOptions)
	if tenant.IsAPINotAvailable(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is not available yet")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	} else if pkgerrors.IsChartConfigNotAvailable(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "chartconfig CRs are not available")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	} else if pkgerrors.IsChartConfigNotInstalled(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "chartconfig CRD does not exist")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	for _, chartSpec := range chartSpecsToMigrate {
		if chartSpec.UserConfigMapName != "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding out if user configmap %#q has been migrated", chartSpec.UserConfigMapName))

			_, err = getChartConfigByName(chartConfigs.Items, chartSpec.ChartName)
			if IsNotFound(err) {
				// We delete the chartconfig CR once the migration process is
				// complete. So if its gone we should not copy the user values
				// again.
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("user configmap %#q has already been migrated", chartSpec.UserConfigMapName))
				continue
			}

			cm, err := getConfigMapByName(clusterConfigMaps.Items, chartSpec.UserConfigMapName)
			if IsNotFound(err) {
				// Copy user configmap from the tenant cluster to the cluster namespace.
				err = r.copyUserConfigMap(ctx, cc.Client.TenantCluster.K8s, cr, chartSpec)
				if err != nil {
					return microerror.Mask(err)
				}
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("user configmap %#q has been migrated", chartSpec.UserConfigMapName))
			} else if cm.Name == chartSpec.UserConfigMapName {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("user configmap %#q has already been migrated", chartSpec.UserConfigMapName))
			} else if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	return nil
}

func (r *Resource) copyUserConfigMap(ctx context.Context, tenantK8sClient kubernetes.Interface, cr v1alpha1.ClusterGuestConfig, chartSpec key.ChartSpec) error {
	currentCM, err := tenantK8sClient.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(chartSpec.UserConfigMapName, metav1.GetOptions{})
	if IsNotFound(err) || len(currentCM.Data) == 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("user configmap %#q has no data to migrate", chartSpec.UserConfigMapName))
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating user configmap %#q in namespace %#q", chartSpec.UserConfigMapName, key.ClusterID(cr)))

	userValues := map[string]string{}

	for k, v := range currentCM.Data {
		// This should not happen if the configmap is correct but is a failsafe
		// to prevent nested configmap data being copied to the new location.
		if strings.Contains(v, "apiVersion: v1") || strings.Contains(v, "kind: ConfigMap") {
			return microerror.Maskf(executionFailedError, "user configmap %#q has invalid data %#q", chartSpec.UserConfigMapName, v)
		}

		userValues[k] = v
	}

	// User configmaps for chartconfig CRs only have keys and values under the
	// configmap block. This needs to be converted to YAML for app CRs.
	values := map[string]interface{}{
		"configmap": userValues,
	}

	yamlValues, err := yaml.Marshal(values)
	if err != nil {
		return microerror.Mask(err)
	}

	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      chartSpec.UserConfigMapName,
			Namespace: key.ClusterID(cr),
		},
		Data: map[string]string{
			"values": string(yamlValues),
		},
	}

	_, err = r.k8sClient.CoreV1().ConfigMaps(key.ClusterID(cr)).Create(&cm)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created user configmap %#q in namespace %#q", chartSpec.UserConfigMapName, key.ClusterID(cr)))

	return nil
}

func (r *Resource) newChartSpecsToMigrate() []key.ChartSpec {
	chartSpecs := []key.ChartSpec{}

	switch r.provider {
	case "aws":
		chartSpecs = append(key.CommonChartSpecs(), awskey.ChartSpecs()...)
	case "azure":
		chartSpecs = append(key.CommonChartSpecs(), azurekey.ChartSpecs()...)
	case "kvm":
		chartSpecs = append(key.CommonChartSpecs(), kvmkey.ChartSpecs()...)
	default:
		chartSpecs = key.CommonChartSpecs()
	}

	chartSpecsToMigrate := []key.ChartSpec{}

	for _, spec := range chartSpecs {
		chartSpecsToMigrate = append(chartSpecsToMigrate, spec)
	}

	return chartSpecsToMigrate
}
