package machinedeployment

import (
	"context"
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"k8s.io/apimachinery/pkg/types"
	capiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/service/controller/key"
	"github.com/giantswarm/cluster-operator/service/internal/object"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
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
	awsCluster, err := a.getAWSCluster(ctx, obj)
	if err != nil {
		return "", microerror.Mask(err)
	}

	apiEndpoint := fmt.Sprintf("api.%s.k8s.%s", key.ClusterID(awsCluster), awsCluster.Spec.Cluster.DNS.Domain)
	return apiEndpoint, nil
}

func (a *accessor) getAWSCluster(ctx context.Context, obj interface{}) (*infrastructurev1alpha2.AWSCluster, error) {
	cr, err := key.ToMachineDeployment(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	cache := object.CacheFromContext(ctx)

	var awsCluster *infrastructurev1alpha2.AWSCluster
	{
		o, exists := cache.Get(clusterCacheKey(cr))
		if exists {
			awsCluster, err = toAWSCluster(o)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		} else {
			nsName := types.NamespacedName{
				Name:      key.ClusterID(&cr),
				Namespace: cr.Namespace,
			}

			err = a.ctrlClient.Get(ctx, nsName, awsCluster)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			cache.Put(clusterCacheKey(cr), awsCluster)
		}
	}

	return awsCluster, nil
}

func clusterCacheKey(cr capiv1alpha2.MachineDeployment) string {
	return fmt.Sprintf("infrastructurev1alpha2.AWSCluster/%s", key.ClusterID(&cr))
}

func toAWSCluster(obj interface{}) (*infrastructurev1alpha2.AWSCluster, error) {
	cluster, ok := obj.(*infrastructurev1alpha2.AWSCluster)
	if !ok {
		return nil, microerror.Mask(wrongTypeError)
	}

	return cluster, nil
}
