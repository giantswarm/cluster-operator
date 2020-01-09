package service

import (
	"context"
	"net"
	"sync"
	"time"

	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/k8sclient/k8srestconfig"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/cluster-operator/flag"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/aws"
	"github.com/giantswarm/cluster-operator/service/controller/azure"
	"github.com/giantswarm/cluster-operator/service/controller/kvm"
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
}

// Service is a type providing implementation of microkit service interface.
type Service struct {
	Version *version.Service

	awsLegacyClusterController   *aws.LegacyCluster
	azureLegacyClusterController *azure.LegacyCluster
	bootOnce                     sync.Once
	kvmLegacyClusterController   *kvm.LegacyCluster
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

	calicoCIDR := config.Viper.GetString(config.Flag.Guest.Cluster.Calico.CIDR)
	clusterIPRange := config.Viper.GetString(config.Flag.Guest.Cluster.Kubernetes.API.ClusterIPRange)
	registryDomain := config.Viper.GetString(config.Flag.Service.Image.Registry.Domain)
	resourceNamespace := config.Viper.GetString(config.Flag.Service.KubeConfig.Secret.Namespace)
	provider := config.Viper.GetString(config.Flag.Service.Provider.Kind)

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

	var k8sClient k8sclient.Interface
	{
		c := k8sclient.ClientsConfig{
			Logger: config.Logger,
			SchemeBuilder: k8sclient.SchemeBuilder{
				corev1alpha1.AddToScheme,
			},

			RestConfig: restConfig,
		}

		k8sClient, err = k8sclient.NewClients(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var apprClient *apprclient.Client
	{
		c := apprclient.Config{
			Fs:     afero.NewOsFs(),
			Logger: config.Logger,

			Address:      defaultCNRAddress,
			Organization: defaultCNROrganization,
		}

		apprClient, err = apprclient.New(c)
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
			// TODO: Reduce the max wait to reduce delay when processing
			// broken tenant clusters.
			//
			//     https://github.com/giantswarm/giantswarm/issues/6703
			//
			// EnsureTillerInstalledMaxWait: 2 * time.Minute,
		}

		tenantCluster, err = tenantcluster.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var awsLegacyClusterController *aws.LegacyCluster
	{
		baseClusterConfig, err := newBaseClusterConfig(config.Flag, config.Viper)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		c := aws.LegacyClusterConfig{
			ApprClient:        apprClient,
			BaseClusterConfig: baseClusterConfig,
			CertSearcher:      certsSearcher,
			Fs:                afero.NewOsFs(),
			K8sClient:         k8sClient,
			Logger:            config.Logger,
			Tenant:            tenantCluster,

			CalicoCIDR:        calicoCIDR,
			ClusterIPRange:    clusterIPRange,
			ProjectName:       project.Name(),
			RegistryDomain:    registryDomain,
			Provider:          provider,
			ResourceNamespace: resourceNamespace,
		}

		awsLegacyClusterController, err = aws.NewLegacyCluster(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var azureLegacyClusterController *azure.LegacyCluster
	{
		baseClusterConfig, err := newBaseClusterConfig(config.Flag, config.Viper)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		c := azure.LegacyClusterConfig{
			ApprClient:        apprClient,
			BaseClusterConfig: baseClusterConfig,
			CertSearcher:      certsSearcher,
			Fs:                afero.NewOsFs(),
			K8sClient:         k8sClient,
			Logger:            config.Logger,
			Tenant:            tenantCluster,

			CalicoCIDR:        calicoCIDR,
			ClusterIPRange:    clusterIPRange,
			ProjectName:       project.Name(),
			Provider:          provider,
			RegistryDomain:    registryDomain,
			ResourceNamespace: resourceNamespace,
		}

		azureLegacyClusterController, err = azure.NewLegacyCluster(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var kvmLegacyClusterController *kvm.LegacyCluster
	{
		baseClusterConfig, err := newBaseClusterConfig(config.Flag, config.Viper)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		c := kvm.LegacyClusterConfig{
			ApprClient:        apprClient,
			BaseClusterConfig: baseClusterConfig,
			CertSearcher:      certsSearcher,
			Fs:                afero.NewOsFs(),
			K8sClient:         k8sClient,
			Logger:            config.Logger,
			Tenant:            tenantCluster,

			CalicoCIDR:        calicoCIDR,
			ClusterIPRange:    clusterIPRange,
			ProjectName:       project.Name(),
			Provider:          provider,
			RegistryDomain:    registryDomain,
			ResourceNamespace: resourceNamespace,
		}

		kvmLegacyClusterController, err = kvm.NewLegacyCluster(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var versionService *version.Service
	{
		versionConfig := version.Config{
			Description:    project.Description(),
			GitCommit:      project.GitSHA(),
			Name:           project.Name(),
			Source:         project.Source(),
			Version:        project.Version(),
			VersionBundles: project.NewVersionBundles(),
		}

		versionService, err = version.New(versionConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s := &Service{
		Version: versionService,

		awsLegacyClusterController:   awsLegacyClusterController,
		bootOnce:                     sync.Once{},
		azureLegacyClusterController: azureLegacyClusterController,
		kvmLegacyClusterController:   kvmLegacyClusterController,
	}

	return s, nil
}

// Boot starts top level service implementation.
func (s *Service) Boot(ctx context.Context) {
	s.bootOnce.Do(func() {
		// Start the controllers.
		go s.awsLegacyClusterController.Boot(ctx)
		go s.azureLegacyClusterController.Boot(ctx)
		go s.kvmLegacyClusterController.Boot(ctx)
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
