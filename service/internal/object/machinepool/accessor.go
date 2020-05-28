package machinepool

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/apimachinery/pkg/types"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	expcapiv1alpha3 "sigs.k8s.io/cluster-api/exp/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/service/controller/key"
	"github.com/giantswarm/cluster-operator/service/internal/object"
)

type accessor struct {
	ctrlClient client.Client
	logger     micrologger.Logger
}

type Config struct {
	CtrlClient client.Client
	Logger     micrologger.Logger
}

func NewAccessor(config Config) (object.Accessor, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	a := &accessor{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}
	return a, nil
}

func (a *accessor) GetAPIEndpoint(ctx context.Context, obj interface{}) (string, error) {
	cluster, err := a.getCluster(ctx, obj)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return cluster.Spec.ControlPlaneEndpoint.Host, nil
}

func (a *accessor) getCluster(ctx context.Context, obj interface{}) (*capiv1alpha3.Cluster, error) {
	cr, err := key.ToMachinePool(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	cache := object.CacheFromContext(ctx)

	var cluster *capiv1alpha3.Cluster
	{
		o, exists := cache.Get(clusterCacheKey(cr))
		if exists {
			cluster, err = toCluster(o)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		} else {
			nsName := types.NamespacedName{
				Name:      key.ClusterID(&cr),
				Namespace: cr.Namespace,
			}

			err = a.ctrlClient.Get(ctx, nsName, cluster)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			cache.Put(clusterCacheKey(cr), cluster)
		}
	}

	return cluster, nil
}

func clusterCacheKey(cr expcapiv1alpha3.MachinePool) string {
	return fmt.Sprintf("capiv1alpha3.Cluster/%s", key.ClusterID(&cr))
}

func toCluster(obj interface{}) (*capiv1alpha3.Cluster, error) {
	cluster, ok := obj.(*capiv1alpha3.Cluster)
	if !ok {
		return nil, microerror.Mask(wrongTypeError)
	}

	return cluster, nil
}
