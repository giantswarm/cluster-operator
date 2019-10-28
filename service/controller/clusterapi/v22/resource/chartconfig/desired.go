package chartconfig

import (
	"context"
	"strconv"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/cluster-operator/pkg/annotation"
	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	pkgkey "github.com/giantswarm/cluster-operator/pkg/v22/key"
	awskey "github.com/giantswarm/cluster-operator/service/controller/aws/v22/key"
	azurekey "github.com/giantswarm/cluster-operator/service/controller/azure/v22/key"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v22/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v22/key"
	kvmkey "github.com/giantswarm/cluster-operator/service/controller/kvm/v22/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var chartConfigs []*g8sv1alpha1.ChartConfig

	for _, chartSpec := range r.newChartSpecs() {
		// No ChartConfig CR is added because it has been migrated to use an
		// App CR.
		if chartSpec.HasAppCR {
			continue
		}

		// App config maps are statically defined by cluster-operator. We put app
		// config maps with user config maps together into the ChartConfig CRs below
		// so they can be merged and take effect accordingly.
		acmSpec, err := r.newConfigMapSpec(*cc, cr, chartSpec.ConfigMapName, chartSpec.Namespace)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		// User config maps are created by cluster-operator and modified by users to
		// override chart values on demand. We put user config maps with app config
		// maps together into the ChartConfig CRs below so they can be merged and
		// take effect accordingly.
		ucmSpec, err := r.newConfigMapSpec(*cc, cr, chartSpec.UserConfigMapName, chartSpec.Namespace)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		chartConfigs = append(chartConfigs, r.newChartConfig(cr, chartSpec, acmSpec, ucmSpec))
	}

	return chartConfigs, nil
}

func (r *Resource) newChartConfig(cr cmav1alpha1.Cluster, chartSpec pkgkey.ChartSpec, acmSpec g8sv1alpha1.ChartConfigSpecConfigMap, ucmSpec g8sv1alpha1.ChartConfigSpecConfigMap) *g8sv1alpha1.ChartConfig {
	return &g8sv1alpha1.ChartConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ChartConfig",
			APIVersion: "core.giantswarm.io",
		},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				annotation.ForceHelmUpgrade: strconv.FormatBool(chartSpec.UseUpgradeForce),
			},
			Labels: map[string]string{
				label.App:          chartSpec.AppName,
				label.Cluster:      key.ClusterID(&cr),
				label.ManagedBy:    project.Name(),
				label.Organization: key.OrganizationID(&cr),
				label.ServiceType:  label.ServiceTypeManaged,
			},
			Name:      chartSpec.ChartName,
			Namespace: "giantswarm",
		},
		Spec: g8sv1alpha1.ChartConfigSpec{
			Chart: g8sv1alpha1.ChartConfigSpecChart{
				Name:      chartSpec.ChartName,
				Namespace: chartSpec.Namespace,
				Channel:   chartSpec.ChannelName,
				Release:   chartSpec.ReleaseName,

				ConfigMap:     acmSpec,
				UserConfigMap: ucmSpec,
			},
			VersionBundle: g8sv1alpha1.ChartConfigSpecVersionBundle{
				Version: "0.6.0",
			},
		},
	}
}

func (r *Resource) newChartSpecs() []pkgkey.ChartSpec {
	switch r.provider {
	case "aws":
		return append(pkgkey.CommonChartSpecs(), awskey.ChartSpecs()...)
	case "azure":
		return append(pkgkey.CommonChartSpecs(), azurekey.ChartSpecs()...)
	case "kvm":
		return append(pkgkey.CommonChartSpecs(), kvmkey.ChartSpecs()...)
	default:
		return pkgkey.CommonChartSpecs()
	}
}

func (r *Resource) newConfigMapSpec(cc controllercontext.Context, cr cmav1alpha1.Cluster, name string, namespace string) (g8sv1alpha1.ChartConfigSpecConfigMap, error) {
	if name == "" {
		// Not all charts define a user config map. In case there is none given, we
		// just retrun the zero value since this is the best we can do.
		return g8sv1alpha1.ChartConfigSpecConfigMap{}, nil
	}

	configMapSpec := g8sv1alpha1.ChartConfigSpecConfigMap{
		Name:      name,
		Namespace: namespace,
	}

	cm, err := cc.Client.TenantCluster.K8s.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) || tenant.IsAPINotAvailable(err) {
		// Cannot get configmap resource version so leave it unset. We will check
		// again after the next resync period.
		return configMapSpec, nil
	} else if err != nil {
		return g8sv1alpha1.ChartConfigSpecConfigMap{}, microerror.Mask(err)
	}

	// Set the configmap resource version. When this changes it will generate an
	// update event for chart-operator. chart-operator will recalculate the
	// desired state including any updated config map values.
	configMapSpec.ResourceVersion = cm.ResourceVersion

	return configMapSpec, nil
}
