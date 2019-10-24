package configmap

import (
	"context"
	"reflect"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/pkg/v21/key"
	awskey "github.com/giantswarm/cluster-operator/service/controller/aws/v21/key"
	azurekey "github.com/giantswarm/cluster-operator/service/controller/azure/v21/key"
	kvmkey "github.com/giantswarm/cluster-operator/service/controller/kvm/v21/key"
)

// Config represents the configuration used to create a new configmap service.
type Config struct {
	Logger micrologger.Logger
	Tenant tenantcluster.Interface

	Provider string
}

// Service provides shared functionality for managing configmaps.
type Service struct {
	logger micrologger.Logger
	tenant tenantcluster.Interface

	provider string
}

// New creates a new configmap service.
func New(config Config) (*Service, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Tenant == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Tenant must not be empty", config)
	}

	if config.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	s := &Service{
		logger: config.Logger,
		tenant: config.Tenant,

		provider: config.Provider,
	}

	return s, nil
}

func (s *Service) newChartSpecs() []key.ChartSpec {
	switch s.provider {
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

func (s *Service) getChartSpecByAppName(appName string) key.ChartSpec {
	for _, spec := range s.newChartSpecs() {
		if spec.AppName == appName {
			return spec
		}
	}

	return key.ChartSpec{}
}

func (s *Service) newTenantK8sClient(ctx context.Context, clusterConfig ClusterConfig) (kubernetes.Interface, error) {
	tenantK8sClient, err := s.tenant.NewK8sClient(ctx, clusterConfig.ClusterID, clusterConfig.APIDomain)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return tenantK8sClient, nil
}

// containsConfigMap checks if item is present within list
// by comparing ObjectMeta Name and Namespace property between item and list objects.
func containsConfigMap(list []*corev1.ConfigMap, item *corev1.ConfigMap) bool {
	for _, l := range list {
		if item.Name == l.Name && item.Namespace == l.Namespace {
			return true
		}
	}

	return false
}

func getConfigMapByNameAndNamespace(list []*corev1.ConfigMap, name, namespace string) (*corev1.ConfigMap, error) {
	for _, l := range list {
		if l.Name == name && l.Namespace == namespace {
			return l, nil
		}
	}

	return nil, microerror.Mask(notFoundError)
}

func isConfigMapModified(a, b *corev1.ConfigMap) bool {
	// If the Data section has changed we need to update.
	if !reflect.DeepEqual(a.Data, b.Data) {
		return true
	}
	// If the Labels have changed we also need to update.
	if !reflect.DeepEqual(a.Labels, b.Labels) {
		return true
	}

	return false
}
