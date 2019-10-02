package chartconfig

import (
	"context"
	"reflect"
	"strings"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/pkg/annotation"
	"github.com/giantswarm/cluster-operator/pkg/v21/key"
	awskey "github.com/giantswarm/cluster-operator/service/controller/aws/v21/key"
	azurekey "github.com/giantswarm/cluster-operator/service/controller/azure/v21/key"
	kvmkey "github.com/giantswarm/cluster-operator/service/controller/kvm/v21/key"
)

const (
	// resourceNamespace is the namespace where the chartconfig CRs are created.
	resourceNamespace = "giantswarm"
)

// Config represents the configuration used to create a new chartconfig service.
type Config struct {
	Logger micrologger.Logger
	Tenant tenantcluster.Interface

	Provider string
}

// ChartConfig provides shared functionality for managing chartconfigs.
type ChartConfig struct {
	logger micrologger.Logger
	tenant tenantcluster.Interface

	provider string
}

// New creates a new chartconfig service.
func New(config Config) (*ChartConfig, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Tenant == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Tenant must not be empty", config)
	}

	if config.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	s := &ChartConfig{
		logger: config.Logger,
		tenant: config.Tenant,

		provider: config.Provider,
	}

	return s, nil
}

func (c *ChartConfig) newChartSpecs() []key.ChartSpec {
	switch c.provider {
	case "aws":
		return append(key.CommonChartSpecs(), awskey.ChartSpecs()...)
	case "azure":
		return append(key.CommonChartSpecs(), azurekey.ChartSpecs()...)
	case "kvm":
		return append(key.CommonChartSpecs(), kvmkey.ChartSpecs()...)
	default:
		return key.CommonChartSpecs()
	}
}

func (c *ChartConfig) getChartSpecByName(name string) key.ChartSpec {
	for _, spec := range c.newChartSpecs() {
		if spec.ChartName == name {
			return spec
		}
	}

	return key.ChartSpec{}
}

func (c *ChartConfig) newTenantG8sClient(ctx context.Context, clusterConfig ClusterConfig) (versioned.Interface, error) {
	tenantG8sClient, err := c.tenant.NewG8sClient(ctx, clusterConfig.ClusterID, clusterConfig.APIDomain)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return tenantG8sClient, nil
}

func (c *ChartConfig) newTenantK8sClient(ctx context.Context, clusterConfig ClusterConfig) (kubernetes.Interface, error) {
	tenantK8sClient, err := c.tenant.NewK8sClient(ctx, clusterConfig.ClusterID, clusterConfig.APIDomain)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return tenantK8sClient, nil
}

// containsChartConfig checks if item is present within list
// by comparing ChartConfig.Name property between item and list objects
// and comparing list object namespace against resourceNamespace
// which is the destination namespace for the item ChartConfig see ApplyCreateChange.
func containsChartConfig(list []*v1alpha1.ChartConfig, item *v1alpha1.ChartConfig) bool {
	for _, l := range list {
		if item.Name == l.Name && l.Namespace == resourceNamespace {
			return true
		}
	}

	return false
}

func getChartConfigByName(list []*v1alpha1.ChartConfig, name string) (*v1alpha1.ChartConfig, error) {
	for _, l := range list {
		if l.Name == name {
			return l, nil
		}
	}

	return nil, microerror.Mask(notFoundError)
}

func filterChartOperatorAnnotations(cr *v1alpha1.ChartConfig) map[string]string {
	annotations := map[string]string{}

	for k, v := range cr.Annotations {
		if k == annotation.CordonReason || k == annotation.CordonUntilDate {
			continue
		}
		if strings.HasPrefix(k, annotation.ChartOperator) {
			annotations[k] = v
		}
	}

	return annotations
}

func isChartConfigModified(a, b *v1alpha1.ChartConfig) bool {
	// If the Spec section has changed we need to update.
	if !reflect.DeepEqual(a.Spec, b.Spec) {
		return true
	}
	// If the Labels have changed we also need to update.
	if !reflect.DeepEqual(a.Labels, b.Labels) {
		return true
	}

	// We only consider annotations with the chart-operator prefix.
	filteredA := filterChartOperatorAnnotations(a)
	filteredB := filterChartOperatorAnnotations(b)

	if !reflect.DeepEqual(filteredA, filteredB) {
		return true
	}

	return false
}
