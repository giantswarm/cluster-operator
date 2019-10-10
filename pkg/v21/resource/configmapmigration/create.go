package configmapmigration

import (
	"context"
	"fmt"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/pkg/v21/key"
	awskey "github.com/giantswarm/cluster-operator/service/controller/aws/v21/key"
	azurekey "github.com/giantswarm/cluster-operator/service/controller/azure/v21/key"
	kvmkey "github.com/giantswarm/cluster-operator/service/controller/kvm/v21/key"
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
		resourcecanceledcontext.SetCanceled(ctx)
		return nil
	}

	chartSpecsToMigrate := r.newChartSpecsToMigrate()

	if len(chartSpecsToMigrate) == 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no charts to migrate")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil
	}

	clusterConfig, err := r.getClusterConfigFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	tenantAPIDomain, err := key.APIDomain(clusterConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	tenantG8sClient, err := r.tenant.NewG8sClient(ctx, clusterConfig.ID, tenantAPIDomain)
	if err != nil {
		return microerror.Mask(err)
	}

	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", label.ManagedBy, project.Name()),
	}

	chartConfigs, err := tenantG8sClient.CoreV1alpha1().ChartConfigs("giantswarm").List(listOptions)
	if tenant.IsAPINotAvailable(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is not available yet")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil
	} else if isChartConfigNotInstalled(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "chartconfig CRD does not exist")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	tenantK8sClient, err := r.tenant.NewK8sClient(ctx, clusterConfig.ID, tenantAPIDomain)
	if err != nil {
		return microerror.Mask(err)
	}

	tenantConfigMaps, err := tenantK8sClient.CoreV1().ConfigMaps(metav1.NamespaceSystem).List(listOptions)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, chartSpec := range chartSpecsToMigrate {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding out if chartconfig %#q has been deleted", chartSpec.ChartName))

		_, err = getChartConfigByName(chartConfigs.Items, chartSpec.ChartName)
		if IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("chartconfig %#q has been deleted", chartSpec.ChartName))
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensuring tenant configmaps are deleted for app %#q", chartSpec.AppName))

			err = r.ensureTenantConfigMapsDeleted(ctx, tenantK8sClient, chartSpec, tenantConfigMaps.Items)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensured tenant configmaps are deleted for app %#q", chartSpec.AppName))

		} else if err == nil {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("chartconfig %#q has not been deleted, continuing", chartSpec.UserConfigMapName))
		}
	}

	return nil
}

func (r *Resource) copyUserConfigMap(ctx context.Context, tenantK8sClient kubernetes.Interface, chartSpec key.ChartSpec) error {
	tenantConfigMap, err := tenantK8sClient.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(chartSpec.ConfigMapName, metav1.GetOptions{})
	if IsNotFound(err) || len(tenantConfigMap.Data) == 0 {
		// Nothing to do.
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) ensureTenantConfigMapsDeleted(ctx context.Context, tenantK8sClient kubernetes.Interface, chartSpec key.ChartSpec, tenantConfigMaps []corev1.ConfigMap) error {
	_, err := getConfigMapByName(tenantConfigMaps, chartSpec.ConfigMapName)
	if IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting tenant configmap %#q", chartSpec.ConfigMapName))

		err = tenantK8sClient.CoreV1().ConfigMaps(metav1.NamespaceSystem).Delete(chartSpec.ConfigMapName, &metav1.DeleteOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted tenant configmap %#q", chartSpec.ConfigMapName))
	}

	_, err = getConfigMapByName(tenantConfigMaps, chartSpec.UserConfigMapName)
	if IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting tenant configmap %#q", chartSpec.UserConfigMapName))

		err = tenantK8sClient.CoreV1().ConfigMaps(metav1.NamespaceSystem).Delete(chartSpec.UserConfigMapName, &metav1.DeleteOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted tenant configmap %#q", chartSpec.UserConfigMapName))
	}

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
		if spec.HasAppCR {
			chartSpecsToMigrate = append(chartSpecsToMigrate, spec)
		}
	}

	return chartSpecsToMigrate
}
