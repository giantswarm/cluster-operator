package service

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/cluster-operator/flag"
	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/service/collector"
	"github.com/giantswarm/cluster-operator/service/controller/aws"
	"github.com/giantswarm/cluster-operator/service/controller/azure"
	"github.com/giantswarm/cluster-operator/service/controller/kvm"
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
}

// Service is a type providing implementation of microkit service interface.
type Service struct {
	Version *version.Service

	awsClusterController   *aws.Cluster
	azureClusterController *azure.Cluster
	bootOnce               sync.Once
	kvmClusterController   *kvm.Cluster
	metricsCollector       *collector.Collector
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

	registryDomain := config.Viper.GetString(config.Flag.Service.Image.Registry.Domain)
	resourceNamespace := config.Viper.GetString(config.Flag.Service.KubeConfig.Secret.Namespace)
	clusterIPRange := config.Viper.GetString(config.Flag.Guest.Cluster.Kubernetes.API.ClusterIPRange)
	calicoAddress := config.Viper.GetString(config.Flag.Guest.Cluster.Calico.Subnet)
	calicoPrefixLength := config.Viper.GetString(config.Flag.Guest.Cluster.Calico.CIDR)

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

	g8sClient, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	k8sExtClient, err := apiextensionsclient.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	fs := afero.NewOsFs()
	var apprClient *apprclient.Client
	{
		c := apprclient.Config{
			Fs:     fs,
			Logger: config.Logger,

			Address:      defaultCNRAddress,
			Organization: defaultCNROrganization,
		}

		apprClient, err = apprclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var certSearcher certs.Interface
	{
		c := certs.Config{
			K8sClient: k8sClient,
			Logger:    config.Logger,

			WatchTimeout: 5 * time.Second,
		}

		certSearcher, err = certs.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var awsClusterController *aws.Cluster
	{
		baseClusterConfig, err := newBaseClusterConfig(config.Flag, config.Viper)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		c := aws.ClusterConfig{
			ApprClient:        apprClient,
			BaseClusterConfig: baseClusterConfig,
			CertSearcher:      certSearcher,
			Fs:                fs,
			G8sClient:         g8sClient,
			K8sClient:         k8sClient,
			K8sExtClient:      k8sExtClient,
			Logger:            config.Logger,

			ClusterIPRange:     clusterIPRange,
			CalicoAddress:      calicoAddress,
			CalicoPrefixLength: calicoPrefixLength,
			ProjectName:        config.ProjectName,
			RegistryDomain:     registryDomain,
			ResourceNamespace:  resourceNamespace,
		}

		awsClusterController, err = aws.NewCluster(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var azureClusterController *azure.Cluster
	{
		baseClusterConfig, err := newBaseClusterConfig(config.Flag, config.Viper)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		c := azure.ClusterConfig{
			ApprClient:        apprClient,
			BaseClusterConfig: baseClusterConfig,
			CertSearcher:      certSearcher,
			Fs:                fs,
			G8sClient:         g8sClient,
			K8sClient:         k8sClient,
			K8sExtClient:      k8sExtClient,
			Logger:            config.Logger,

			ClusterIPRange:     clusterIPRange,
			CalicoAddress:      calicoAddress,
			CalicoPrefixLength: calicoPrefixLength,
			ProjectName:        config.ProjectName,
			RegistryDomain:     registryDomain,
			ResourceNamespace:  resourceNamespace,
		}

		azureClusterController, err = azure.NewCluster(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var kvmClusterController *kvm.Cluster
	{
		baseClusterConfig, err := newBaseClusterConfig(config.Flag, config.Viper)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		c := kvm.ClusterConfig{
			ApprClient:        apprClient,
			BaseClusterConfig: baseClusterConfig,
			CertSearcher:      certSearcher,
			Fs:                fs,
			G8sClient:         g8sClient,
			K8sClient:         k8sClient,
			K8sExtClient:      k8sExtClient,
			Logger:            config.Logger,

			ClusterIPRange:     clusterIPRange,
			CalicoAddress:      calicoAddress,
			CalicoPrefixLength: calicoPrefixLength,
			ProjectName:        config.ProjectName,
			RegistryDomain:     registryDomain,
			ResourceNamespace:  resourceNamespace,
		}

		kvmClusterController, err = kvm.NewCluster(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var metricsCollector *collector.Collector
	{
		c := collector.Config{
			CertSearcher: certSearcher,
			G8sClient:    g8sClient,
			Logger:       config.Logger,
		}

		metricsCollector, err = collector.New(c)
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
			VersionBundles: NewVersionBundles(),
		}

		versionService, err = version.New(versionConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s := &Service{
		Version: versionService,

		awsClusterController:   awsClusterController,
		bootOnce:               sync.Once{},
		azureClusterController: azureClusterController,
		kvmClusterController:   kvmClusterController,
		metricsCollector:       metricsCollector,
	}

	return s, nil
}

// Boot starts top level service implementation.
func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		prometheus.MustRegister(s.metricsCollector)

		// Start the controllers.
		go s.awsClusterController.Boot(context.Background())
		go s.azureClusterController.Boot(context.Background())
		go s.kvmClusterController.Boot(context.Background())
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
