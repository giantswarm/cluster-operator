package clusterconfigmap

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/v5/service/internal/basedomain"
	"github.com/giantswarm/cluster-operator/v5/service/internal/podcidr"
)

const (
	// Name is the identifier of the resource.
	Name = "clusterconfigmap"
)

// Config represents the configuration used to create a new clusterConfigMap
// resource.
type Config struct {
	BaseDomain basedomain.Interface
	CtrlClient ctrlClient.Client
	K8sClient  kubernetes.Interface
	Logger     micrologger.Logger
	PodCIDR    podcidr.Interface

	ClusterIPRange string
	DNSIP          string
	Installation   string
	Provider       string
}

// Resource implements the clusterConfigMap resource.
type Resource struct {
	baseDomain basedomain.Interface
	ctrlClient ctrlClient.Client
	k8sClient  kubernetes.Interface
	logger     micrologger.Logger
	podCIDR    podcidr.Interface

	clusterIPRange string
	dnsIP          string
	provider       string
}

// New creates a new configured config map state getter resource managing
// cluster config maps.
//
//	https://pkg.go.dev/github.com/giantswarm/operatorkit/v8/pkg/resource/k8s/secretresource#StateGetter
func New(config Config) (*Resource, error) {
	if config.BaseDomain == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.BaseDomain must not be empty", config)
	}
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.PodCIDR == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.PodCIDR must not be empty", config)
	}

	if config.ClusterIPRange == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterIPRange must not be empty", config)
	}
	if config.DNSIP == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.DNSIP must not be empty", config)
	}
	if config.Installation == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Installation must not be empty", config)
	}
	if config.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	r := &Resource{
		baseDomain: config.BaseDomain,
		ctrlClient: config.CtrlClient,
		k8sClient:  config.K8sClient,
		logger:     config.Logger,
		podCIDR:    config.PodCIDR,

		clusterIPRange: config.ClusterIPRange,
		dnsIP:          config.DNSIP,
		provider:       config.Provider,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
