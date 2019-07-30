package service

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/clusterclient"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	"github.com/giantswarm/tenantcluster"
	"github.com/go-resty/resty"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/cluster-operator/flag"
	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/service/collector"
	"github.com/giantswarm/cluster-operator/service/controller/aws"
	"github.com/giantswarm/cluster-operator/service/controller/azure"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi"
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
	Version     string
}

// Service is a type providing implementation of microkit service interface.
type Service struct {
	Version *version.Service

	awsLegacyClusterController   *aws.LegacyCluster
	azureLegacyClusterController *azure.LegacyCluster
	bootOnce                     sync.Once
	clusterController            *clusterapi.Cluster
	machineDeploymentController  *clusterapi.MachineDeployment
	kvmLegacyClusterController   *kvm.LegacyCluster
	operatorCollector            *collector.Set
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

	cmaClient, err := clientset.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
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

	var tenantCluster tenantcluster.Interface
	{
		c := tenantcluster.Config{
			CertsSearcher: certSearcher,
			Logger:        config.Logger,

			CertID: certs.ClusterOperatorAPICert,
			// This is used by the Tiller resource where we use a shorter max
			// wait because the tenant cluster may be unavailable. If so the
			// reconciliation loop is cancelled and we retry in the next loop.
			EnsureTillerInstalledMaxWait: 45 * time.Second,
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
			CertSearcher:      certSearcher,
			Fs:                fs,
			G8sClient:         g8sClient,
			K8sClient:         k8sClient,
			K8sExtClient:      k8sExtClient,
			Logger:            config.Logger,
			Tenant:            tenantCluster,

			ClusterIPRange:     clusterIPRange,
			CalicoAddress:      calicoAddress,
			CalicoPrefixLength: calicoPrefixLength,
			ProjectName:        config.ProjectName,
			RegistryDomain:     registryDomain,
			ResourceNamespace:  resourceNamespace,
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
			CertSearcher:      certSearcher,
			Fs:                fs,
			G8sClient:         g8sClient,
			K8sClient:         k8sClient,
			K8sExtClient:      k8sExtClient,
			Logger:            config.Logger,
			Tenant:            tenantCluster,

			ClusterIPRange:     clusterIPRange,
			CalicoAddress:      calicoAddress,
			CalicoPrefixLength: calicoPrefixLength,
			ProjectName:        config.ProjectName,
			RegistryDomain:     registryDomain,
			ResourceNamespace:  resourceNamespace,
		}

		azureLegacyClusterController, err = azure.NewLegacyCluster(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterController *clusterapi.Cluster
	{
		baseClusterConfig, err := newBaseClusterConfig(config.Flag, config.Viper)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		var clusterClient *clusterclient.Client
		{
			config.Logger.Log("level", "debug", "message", fmt.Sprintf("address for cluster-service: %q", config.Viper.GetString(config.Flag.Service.ClusterService.Address)))

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

		c := clusterapi.ClusterConfig{
			BaseClusterConfig: baseClusterConfig,
			ClusterClient:     clusterClient,
			CMAClient:         cmaClient,
			G8sClient:         g8sClient,
			K8sExtClient:      k8sExtClient,
			Logger:            config.Logger,
			Tenant:            tenantCluster,

			ProjectName: config.ProjectName,
		}

		clusterController, err = clusterapi.NewCluster(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var machineDeploymentController *clusterapi.MachineDeployment
	{
		c := clusterapi.MachineDeploymentConfig{
			CMAClient:    cmaClient,
			G8sClient:    g8sClient,
			K8sExtClient: k8sExtClient,
			Logger:       config.Logger,
			Tenant:       tenantCluster,

			ProjectName: config.ProjectName,
		}

		machineDeploymentController, err = clusterapi.NewMachineDeployment(c)
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
			CertSearcher:      certSearcher,
			Fs:                fs,
			G8sClient:         g8sClient,
			K8sClient:         k8sClient,
			K8sExtClient:      k8sExtClient,
			Logger:            config.Logger,
			Tenant:            tenantCluster,

			ClusterIPRange:     clusterIPRange,
			CalicoAddress:      calicoAddress,
			CalicoPrefixLength: calicoPrefixLength,
			ProjectName:        config.ProjectName,
			RegistryDomain:     registryDomain,
			ResourceNamespace:  resourceNamespace,
		}

		kvmLegacyClusterController, err = kvm.NewLegacyCluster(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var operatorCollector *collector.Set
	{
		c := collector.SetConfig{
			CertSearcher: certSearcher,
			CMAClient:    cmaClient,
			G8sClient:    g8sClient,
			Logger:       config.Logger,
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
			VersionBundles: NewVersionBundles(),
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
		clusterController:            clusterController,
		machineDeploymentController:  machineDeploymentController,
		kvmLegacyClusterController:   kvmLegacyClusterController,
		operatorCollector:            operatorCollector,
	}

	return s, nil
}

// Boot starts top level service implementation.
func (s *Service) Boot(ctx context.Context) {
	s.bootOnce.Do(func() {
		go s.operatorCollector.Boot(ctx)

		// Start the controllers.
		go s.awsLegacyClusterController.Boot(ctx)
		go s.azureLegacyClusterController.Boot(ctx)
		go s.clusterController.Boot(ctx)
		go s.machineDeploymentController.Boot(ctx)
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
