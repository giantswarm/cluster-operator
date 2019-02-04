package chartconfigcrd

import (
	"context"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/pkg/v10/key"
	"github.com/giantswarm/errors/guest"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
)

// EnsureCreated ensures the chartconfig crd is created in the tenant cluster
// this allows the chartconfig resource to create CRs before chart-operator has
// booted.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	clusterGuestConfig, err := r.toClusterGuestConfigFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	clusterConfig, err := prepareClusterConfig(r.baseClusterConfig, clusterGuestConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	tenantAPIDomain, err := key.APIDomain(clusterGuestConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	tenantK8sExtClient, err := r.tenant.NewK8sExtClient(ctx, clusterConfig.ClusterID, tenantAPIDomain)
	if err != nil {
		return microerror.Mask(err)
	}

	var crdClient *k8scrdclient.CRDClient
	{
		c := k8scrdclient.Config{
			K8sExtClient: tenantK8sExtClient,
			Logger:       r.logger,
		}

		crdClient, err = k8scrdclient.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	crdBackoff := backoff.NewMaxRetries(3, 1*time.Second)
	err = crdClient.EnsureCreated(ctx, v1alpha1.NewChartConfigCRD(), crdBackoff)
	if guest.IsAPINotAvailable(err) {
		// We should not hammer tenant API if it is not available, the tenant cluster
		// might be initializing. We will retry on next reconciliation loop.
		reconciliationcanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")

	} else if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func prepareClusterConfig(baseClusterConfig cluster.Config, clusterGuestConfig v1alpha1.ClusterGuestConfig) (cluster.Config, error) {
	var err error

	// Use baseClusterConfig as a basis and supplement it with information from
	// clusterGuestConfig.
	clusterConfig := baseClusterConfig

	clusterConfig.ClusterID = key.ClusterID(clusterGuestConfig)
	clusterConfig.Domain.API, err = key.APIDomain(clusterGuestConfig)
	if err != nil {
		return cluster.Config{}, microerror.Mask(err)
	}

	return clusterConfig, nil
}
