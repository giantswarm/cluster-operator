package service

import (
	"context"
	"net"
	"sync"
	"time"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/clusterclient"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/k8sclient/k8srestconfig"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster"
	"github.com/giantswarm/versionbundle"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"gopkg.in/resty.v1"
	"k8s.io/client-go/rest"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/cluster-operator/flag"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/collector"
	"github.com/giantswarm/cluster-operator/service/controller"
	"github.com/giantswarm/cluster-operator/service/controller/key"
	"github.com/giantswarm/cluster-operator/service/internal/cluster"
)

const (
	apiServerIPLastOctet = 1

	defaultCNRAddress      = "https://quay.io"
	defaultCNROrganization = "giantswarm"
)

// Config represents the configuration used to create a new service.
type Config struct {
	Logger micrologger.Logger

	Flag  *flag.Flag
	Viper *viper.Viper

	Description string
	GitCommit   string
	ProjectName string
	Source      string
	Version     string
}

// Service is a type providing implementation of microkit service interface.
type Service struct {
	Version *version.Service

	bootOnce                    sync.Once
	clusterController           *controller.Cluster
	machineDeploymentController *controller.MachineDeployment
	operatorCollector           *collector.Set
}

// New creates a new service with given configuration.
func New(config Config) (*Service, error) {
	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Flag must not be empty", config)
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Viper must not be empty", config)
	}

	var err error

	calicoAddress := config.Viper.GetString(config.Flag.Guest.Cluster.Calico.Subnet)
	calicoPrefixLength := config.Viper.GetString(config.Flag.Guest.Cluster.Calico.CIDR)
	clusterIPRange := config.Viper.GetString(config.Flag.Guest.Cluster.Kubernetes.API.ClusterIPRange)
	provider := config.Viper.GetString(config.Flag.Service.Provider.Kind)
	registryDomain := config.Viper.GetString(config.Flag.Service.Image.Registry.Domain)

	var restConfig *rest.Config
	{
		c := k8srestconfig.Config{
			Logger: config.Logger,

			Address:    config.Viper.GetString(config.Flag.Service.Kubernetes.Address),
			InCluster:  config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster),
			KubeConfig: config.Viper.GetString(config.Flag.Service.Kubernetes.KubeConfig),
			TLS: k8srestconfig.ConfigTLS{
				CAFile:  config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile),
				CrtFile: config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile),
				KeyFile: config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile),
			},
		}

		restConfig, err = k8srestconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var k8sClient *k8sclient.Clients
	{
		c := k8sclient.ClientsConfig{
			SchemeBuilder: k8sclient.SchemeBuilder{
				infrastructurev1alpha2.AddToScheme,
				apiv1alpha3.AddToScheme,
			},
			Logger: config.Logger,

			RestConfig: restConfig,
		}

		k8sClient, err = k8sclient.NewClients(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var dnsIP string
	{
		dnsIP, err = key.DNSIP(clusterIPRange)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var apiIP string
	{
		_, ip, err := parseClusterIPRange(config.Viper.GetString(config.Flag.Guest.Cluster.Kubernetes.API.ClusterIPRange))
		if err != nil {
			return nil, microerror.Mask(err)
		}
		apiIP = ip.String()
	}

	var clusterClient *clusterclient.Client
	{
		c := clusterclient.Config{
			Address: config.Viper.GetString(config.Flag.Service.ClusterService.Address),
			Logger:  config.Logger,

			// Timeout & RetryCount are straight from `api/service/service.go`.
			RestClient: resty.New().SetTimeout(15 * time.Second).SetRetryCount(5),
		}

		clusterClient, err = clusterclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var certsSearcher certs.Interface
	{
		c := certs.Config{
			K8sClient: k8sClient.K8sClient(),
			Logger:    config.Logger,

			WatchTimeout: 5 * time.Second,
		}

		certsSearcher, err = certs.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tenantCluster tenantcluster.Interface
	{
		c := tenantcluster.Config{
			CertsSearcher: certsSearcher,
			Logger:        config.Logger,

			CertID: certs.ClusterOperatorAPICert,
		}

		tenantCluster, err = tenantcluster.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterController *controller.Cluster
	{
		c := controller.ClusterConfig{
			CertsSearcher: certsSearcher,
			ClusterClient: clusterClient,
			FileSystem:    afero.NewOsFs(),
			K8sClient:     k8sClient,
			Logger:        config.Logger,
			Tenant:        tenantCluster,

			APIIP:                      apiIP,
			CalicoAddress:              calicoAddress,
			CalicoPrefixLength:         calicoPrefixLength,
			CertTTL:                    config.Viper.GetString(config.Flag.Guest.Cluster.Vault.Certificate.TTL),
			ClusterIPRange:             clusterIPRange,
			DNSIP:                      dnsIP,
			NewCommonClusterObjectFunc: newCommonClusterObjectFunc(provider),
			Provider:                   provider,
			RegistryDomain:             registryDomain,
		}

		clusterController, err = controller.NewCluster(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var machineDeploymentController *controller.MachineDeployment
	{
		c := controller.MachineDeploymentConfig{
			ClusterClient: clusterClient,
			K8sClient:     k8sClient,
			Logger:        config.Logger,
			Tenant:        tenantCluster,

			Provider: provider,
		}

		machineDeploymentController, err = controller.NewMachineDeployment(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var operatorCollector *collector.Set
	{
		c := collector.SetConfig{
			CertSearcher: certsSearcher,
			K8sClient:    k8sClient,
			Logger:       config.Logger,

			NewCommonClusterObjectFunc: newCommonClusterObjectFunc(provider),
		}

		operatorCollector, err = collector.NewSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var versionService *version.Service
	{
		versionConfig := version.Config{
			Description:    config.Description,
			GitCommit:      config.GitCommit,
			Name:           config.ProjectName,
			Source:         config.Source,
			Version:        config.Version,
			VersionBundles: []versionbundle.Bundle{project.VersionBundle(provider)},
		}

		versionService, err = version.New(versionConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s := &Service{
		Version: versionService,

		bootOnce:                    sync.Once{},
		clusterController:           clusterController,
		machineDeploymentController: machineDeploymentController,
		operatorCollector:           operatorCollector,
	}

	return s, nil
}

// Boot starts top level service implementation.
func (s *Service) Boot(ctx context.Context) {
	s.bootOnce.Do(func() {
		go s.operatorCollector.Boot(ctx)

		// Start the controllers.
		go s.clusterController.Boot(ctx)
		go s.machineDeploymentController.Boot(ctx)
	})
}

func newBaseClusterConfig(f *flag.Flag, v *viper.Viper) (*cluster.Config, error) {
	networkIP, apiServerIP, err := parseClusterIPRange(v.GetString(f.Guest.Cluster.Kubernetes.API.ClusterIPRange))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterConfig := &cluster.Config{
		CertTTL: v.GetString(f.Guest.Cluster.Vault.Certificate.TTL),
		IP: cluster.IP{
			API:   apiServerIP,
			Range: networkIP,
		},
	}

	return clusterConfig, nil
}

func newCommonClusterObjectFunc(provider string) func() infrastructurev1alpha2.CommonClusterObject {
	// Deal with different providers in here once they reach Cluster API.
	return func() infrastructurev1alpha2.CommonClusterObject {
		return new(infrastructurev1alpha2.AWSCluster)
	}
}

func parseClusterIPRange(ipRange string) (net.IP, net.IP, error) {
	_, cidr, err := net.ParseCIDR(ipRange)
	if cidr == nil {
		return nil, nil, microerror.Maskf(invalidConfigError, "invalid Kubernetes ClusterIPRange '%s': cidr == nil", ipRange)
	} else if err != nil {
		return nil, nil, microerror.Maskf(invalidConfigError, "invalid Kubernetes ClusterIPRange '%s': %q", ipRange, err)
	}

	ones, bits := cidr.Mask.Size()
	if bits != 32 {
		return nil, nil, microerror.Maskf(invalidConfigError, "Kubernetes ClusterIPRange CIDR must be an IPv4 range")
	}

	// Node gets /24 from Kubernetes and each POD receives one IP from this
	// block. Therefore CIDR block must be at least /24.
	if ones > 24 {
		return nil, nil, microerror.Maskf(invalidConfigError, "Kubernetes ClusterIPRange CIDR network block must be at least /24")
	}

	networkIP := cidr.IP.To4()
	apiServerIP := net.IPv4(networkIP[0], networkIP[1], networkIP[2], apiServerIPLastOctet)

	return networkIP, apiServerIP, nil
}
