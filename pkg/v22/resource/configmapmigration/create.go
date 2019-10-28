package configmapmigration

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/pkg/v21/controllercontext"
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

	cr, err := r.getClusterConfigFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// Get all configmaps in the cluster namespace.
	clusterConfigMaps, err := r.k8sClient.CoreV1().ConfigMaps(key.ClusterID(cr)).List(metav1.ListOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	if cc.Client.TenantCluster.G8s == nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant clients not available")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
		return nil
	}

	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", label.ManagedBy, project.Name()),
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Get all chartconfig CRs in the tenant cluster. The migration needs to
	// complete before we create app CRs. So we cancel the entire loop on error.
	chartConfigs, err := cc.Client.TenantCluster.G8s.CoreV1alpha1().ChartConfigs("giantswarm").List(listOptions)
	if tenant.IsAPINotAvailable(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is not available yet")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
		return nil
	} else if isChartConfigNotInstalled(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "chartconfig CRD does not exist")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "timeout getting chartconfig CRs")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
		return nil
	}

	// Get all configmaps in kube-system in the tenant cluster. The migration needs to
	// complete before we create app CRs. So we cancel the entire loop on error.
	tenantConfigMaps, err := cc.Client.TenantCluster.K8s.CoreV1().ConfigMaps(metav1.NamespaceSystem).List(listOptions)
	if tenant.IsAPINotAvailable(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is not available yet")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "timeout getting chartconfig CRs")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
		return nil
	}

	for _, chartSpec := range chartSpecsToMigrate {
		if chartSpec.UserConfigMapName != "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding out if user configmap %#q has been migrated", chartSpec.UserConfigMapName))

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

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding out if chartconfig %#q has been deleted", chartSpec.ChartName))

		_, err = getChartConfigByName(chartConfigs.Items, chartSpec.ChartName)
		if IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("chartconfig %#q has been deleted", chartSpec.ChartName))
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensuring tenant configmaps are deleted for app %#q", chartSpec.AppName))

			// Once chartconfig CR is deleted also delete old configmaps in the tenant cluster.
			err = r.ensureTenantConfigMapsDeleted(ctx, cc.Client.TenantCluster.K8s, chartSpec, tenantConfigMaps.Items)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensured tenant configmaps are deleted for app %#q", chartSpec.AppName))

		} else if err == nil {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("chartconfig %#q has not been deleted, continuing", chartSpec.ChartName))
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

	// User configmaps for chartconfig CRs only have keys and values under the
	// configmap block. This needs to be converted to YAML for app CRs.
	values := map[string]interface{}{
		"configmap": currentCM.Data,
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

func (r *Resource) ensureTenantConfigMapsDeleted(ctx context.Context, tenantK8sClient kubernetes.Interface, chartSpec key.ChartSpec, tenantConfigMaps []corev1.ConfigMap) error {
	if chartSpec.ConfigMapName != "" {
		err := r.ensureTenantConfigMapDeleted(ctx, tenantK8sClient, chartSpec.ConfigMapName, tenantConfigMaps)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	if chartSpec.UserConfigMapName != "" {
		err := r.ensureTenantConfigMapDeleted(ctx, tenantK8sClient, chartSpec.UserConfigMapName, tenantConfigMaps)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (r *Resource) ensureTenantConfigMapDeleted(ctx context.Context, tenantK8sClient kubernetes.Interface, configMapName string, tenantConfigMaps []corev1.ConfigMap) error {
	cm, err := getConfigMapByName(tenantConfigMaps, configMapName)
	if cm.Name == configMapName {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting tenant configmap %#q", configMapName))

		err = tenantK8sClient.CoreV1().ConfigMaps(metav1.NamespaceSystem).Delete(configMapName, &metav1.DeleteOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted tenant configmap %#q", configMapName))
	} else if IsNotFound(err) {
		// fall through
	} else if err != nil {
		return microerror.Mask(err)
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
