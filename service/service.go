package service

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha3"
	providerv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/certs/v3/pkg/certs"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/k8sclient/v5/pkg/k8srestconfig"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/cluster-operator/v3/flag"
	"github.com/giantswarm/cluster-operator/v3/pkg/label"
	"github.com/giantswarm/cluster-operator/v3/pkg/project"
	"github.com/giantswarm/cluster-operator/v3/service/collector"
	"github.com/giantswarm/cluster-operator/v3/service/controller"
	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
	"github.com/giantswarm/cluster-operator/v3/service/internal/basedomain"
	"github.com/giantswarm/cluster-operator/v3/service/internal/nodecount"
	"github.com/giantswarm/cluster-operator/v3/service/internal/podcidr"
	"github.com/giantswarm/cluster-operator/v3/service/internal/recorder"
	"github.com/giantswarm/cluster-operator/v3/service/internal/releaseversion"
	"github.com/giantswarm/cluster-operator/v3/service/internal/tenantclient"
)

const (
	apiServerIPLastOctet = 1
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

type operatorkitController interface {
	Boot(ctx context.Context)
}

// Service is a type providing implementation of microkit service interface.
type Service struct {
	Version *version.Service

	bootOnce sync.Once

	controllers       []operatorkitController
	operatorCollector *collector.Set
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

	calicoSubnet := config.Viper.GetString(config.Flag.Guest.Cluster.Calico.Subnet)
	calicoCIDR := config.Viper.GetString(config.Flag.Guest.Cluster.Calico.CIDR)
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
		schemeBuilder := k8sclient.SchemeBuilder{}
		schemeBuilder = append(schemeBuilder, capiv1alpha3.AddToScheme)
		schemeBuilder = append(schemeBuilder, releasev1alpha1.AddToScheme)
		schemeBuilder = append(schemeBuilder, providerv1alpha1.AddToScheme)
		switch provider {
		case label.ProviderAWS:
			schemeBuilder = append(schemeBuilder, infrastructurev1alpha3.AddToScheme)
		}

		c := k8sclient.ClientsConfig{
			SchemeBuilder: schemeBuilder,
			Logger:        config.Logger,

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

	var pc podcidr.Interface
	{
		c := podcidr.Config{
			K8sClient: k8sClient,

			InstallationCIDR: fmt.Sprintf("%s/%s", calicoSubnet, calicoCIDR),
			Provider:         provider,
		}

		pc, err = podcidr.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var bd basedomain.Interface
	{
		c := basedomain.Config{
			K8sClient: k8sClient,
			Provider:  provider,
		}

		bd, err = basedomain.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tenantClient tenantclient.Interface
	{
		c := tenantclient.Config{
			K8sClient:     k8sClient,
			BaseDomain:    bd,
			TenantCluster: tenantCluster,
			Logger:        config.Logger,
		}

		tenantClient, err = tenantclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var nc nodecount.Interface
	{
		c := nodecount.Config{
			K8sClient:    k8sClient,
			TenantClient: tenantClient,
		}

		nc, err = nodecount.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var rv releaseversion.Interface
	{
		c := releaseversion.Config{
			K8sClient: k8sClient,
		}

		rv, err = releaseversion.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var eventRecorder recorder.Interface
	{
		c := recorder.Config{
			K8sClient: k8sClient,

			Component: fmt.Sprintf("%s-%s", project.Name(), project.Version()),
		}

		eventRecorder = recorder.New(c)
	}

	var controllers []operatorkitController
	{
		{
			c := controller.ClusterConfig{
				BaseDomain:     bd,
				CertsSearcher:  certsSearcher,
				Event:          eventRecorder,
				FileSystem:     afero.NewOsFs(),
				K8sClient:      k8sClient,
				Logger:         config.Logger,
				PodCIDR:        pc,
				Tenant:         tenantCluster,
				ReleaseVersion: rv,

				APIIP:                      apiIP,
				CertTTL:                    config.Viper.GetString(config.Flag.Guest.Cluster.Vault.Certificate.TTL),
				ClusterIPRange:             clusterIPRange,
				DNSIP:                      dnsIP,
				ClusterDomain:              config.Viper.GetString(config.Flag.Guest.Cluster.Kubernetes.ClusterDomain),
				KiamWatchDogEnabled:        config.Viper.GetBool(config.Flag.Service.Release.App.Config.KiamWatchDogEnabled),
				NewCommonClusterObjectFunc: newCommonClusterObjectFunc(provider),
				Provider:                   provider,
				RawAppDefaultConfig:        config.Viper.GetString(config.Flag.Service.Release.App.Config.Default),
				RawAppOverrideConfig:       config.Viper.GetString(config.Flag.Service.Release.App.Config.Override),
				RegistryDomain:             registryDomain,
			}

			clusterController, err := controller.NewCluster(c)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			controllers = append(controllers, clusterController)
		}

		if provider == label.ProviderAWS {
			c := controller.ControlPlaneConfig{
				BaseDomain:     bd,
				Event:          eventRecorder,
				K8sClient:      k8sClient,
				Logger:         config.Logger,
				NodeCount:      nc,
				Tenant:         tenantCluster,
				ReleaseVersion: rv,

				Provider: provider,
			}

			controlPlaneController, err := controller.NewControlPlane(c)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			controllers = append(controllers, controlPlaneController)
		}

		var machineDeploymentController *controller.MachineDeployment
		{
			c := controller.MachineDeploymentConfig{
				BaseDomain:     bd,
				Event:          eventRecorder,
				K8sClient:      k8sClient,
				Logger:         config.Logger,
				NodeCount:      nc,
				Tenant:         tenantCluster,
				ReleaseVersion: rv,

				Provider: provider,
			}

			machineDeploymentController, err = controller.NewMachineDeployment(c)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			controllers = append(controllers, machineDeploymentController)
		}
	}

	var operatorCollector *collector.Set
	{
		c := collector.SetConfig{
			CertSearcher: certsSearcher,
			K8sClient:    k8sClient,
			Logger:       config.Logger,

			NewCommonClusterObjectFunc: newCommonClusterObjectFunc(provider),
			Provider:                   provider,
		}

		operatorCollector, err = collector.NewSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var versionService *version.Service
	{
		versionConfig := version.Config{
			Description: config.Description,
			GitCommit:   config.GitCommit,
			Name:        config.ProjectName,
			Source:      config.Source,
			Version:     config.Version,
		}

		versionService, err = version.New(versionConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s := &Service{
		Version: versionService,

		bootOnce:          sync.Once{},
		controllers:       controllers,
		operatorCollector: operatorCollector,
	}

	return s, nil
}

// Boot starts top level service implementation.
func (s *Service) Boot(ctx context.Context) {
	s.bootOnce.Do(func() {
		go func() {
			err := s.operatorCollector.Boot(ctx)
			if err != nil {
				panic(microerror.JSON(err))
			}
		}()

		// Start the controllers.
		for _, c := range s.controllers {
			go c.Boot(ctx)
		}
	})
}

func newCommonClusterObjectFunc(provider string) func() infrastructurev1alpha3.CommonClusterObject {
	// Deal with different providers in here once they reach Cluster API.
	return func() infrastructurev1alpha3.CommonClusterObject {
		return new(infrastructurev1alpha3.AWSCluster)
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
