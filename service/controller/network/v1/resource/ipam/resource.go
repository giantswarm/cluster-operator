package ipam

import (
	"context"
	"net"
	"reflect"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/crdstorage"
	"github.com/giantswarm/ipam"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"
	"github.com/giantswarm/microstorage/memory"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "ipamv1"
)

// Config represents the configuration used to create a new cluster network config resource.
type Config struct {
	CRDClient *k8scrdclient.CRDClient
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	// Network defines the network available for guest clusters.
	Network     net.IPNet
	ProjectName string
	// ReservedSubnets is for declaring specific subnets of Network to be
	// reserved for non-guest cluster use. Such example is e.g. in Azure where
	// host cluster shares same network with guests.
	ReservedSubnets []net.IPNet
	StorageKind     string
}

// Resource implements the cluster network config resource.
type Resource struct {
	g8sClient versioned.Interface
	logger    micrologger.Logger

	ipam *ipam.Service
}

// New creates a new configured cluster network IPAM resource instance.
func New(config Config) (*Resource, error) {
	if config.CRDClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CRDClient must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if reflect.DeepEqual(config.Network, net.IPNet{}) {
		return nil, microerror.Maskf(invalidConfigError, "%T.Network must not be empty", config)
	}
	// Empty config.ReservedSubnets is fine.
	// config.StorageKind is validated below.

	var err error
	var storage microstorage.Storage
	{
		switch config.StorageKind {
		case "":
			return nil, microerror.Maskf(invalidConfigError, "%T.StorageKind must not be empty", config)
		case "memory":
			config.Logger.Log("level", "debug", "message", "using memory storage")
			storage, err = memory.New(memory.DefaultConfig())
		case "crd":
			// NOTE: If crdstorage implementation is removed at some point,
			// remember to remove also namespace creation and storageconfig
			// RBAC rules.
			c := crdstorage.DefaultConfig()

			c.CRDClient = config.CRDClient
			c.G8sClient = config.G8sClient
			c.K8sClient = config.K8sClient
			c.Logger = config.Logger

			c.Name = config.ProjectName
			c.Namespace = &v1.Namespace{
				ObjectMeta: apismetav1.ObjectMeta{
					Name: "giantswarm",
				},
			}

			crdStorage, err := crdstorage.New(c)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			config.Logger.Log("level", "info", "message", "booting crdstorage")
			err = crdStorage.Boot(context.TODO())
			if err != nil {
				return nil, microerror.Mask(err)
			}

			storage = crdStorage
		default:
			return nil, microerror.Maskf(invalidConfigError, "unknown %T.StorageKind: %q", config, config.StorageKind)
		}
	}

	var ipamSvc *ipam.Service
	{
		c := ipam.Config{
			Logger:           config.Logger,
			Storage:          storage,
			Network:          &config.Network,
			AllocatedSubnets: config.ReservedSubnets,
		}

		ipamSvc, err = ipam.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newService := &Resource{
		g8sClient: config.G8sClient,
		logger:    config.Logger,
		ipam:      ipamSvc,
	}

	return newService, nil
}

// Name returns name of the Resource.
func (r *Resource) Name() string {
	return Name
}
