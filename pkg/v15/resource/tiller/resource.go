package tiller

import (
	"context"
	"reflect"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"github.com/giantswarm/tenantcluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/pkg/v10/key"
)

const (
	Name = "tillerv15"
)

// Config represents the configuration used to create a new tiller resource.
type Config struct {
	BaseClusterConfig        cluster.Config
	Logger                   micrologger.Logger
	Tenant                   tenantcluster.Interface
	ToClusterGuestConfigFunc func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
	ToClusterObjectMetaFunc  func(obj interface{}) (metav1.ObjectMeta, error)
}

// Resource implements the tiller resource.
type Resource struct {
	baseClusterConfig        cluster.Config
	logger                   micrologger.Logger
	tenant                   tenantcluster.Interface
	toClusterGuestConfigFunc func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
	toClusterObjectMetaFunc  func(obj interface{}) (metav1.ObjectMeta, error)
}

// New creates a new configured tiller resource.
func New(config Config) (*Resource, error) {
	if reflect.DeepEqual(config.BaseClusterConfig, cluster.Config{}) {
		return nil, microerror.Maskf(invalidConfigError, "%T.BaseClusterConfig must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Tenant == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Tenant must not be empty", config)
	}
	if config.ToClusterGuestConfigFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterGuestConfigFunc must not be empty", config)
	}
	if config.ToClusterObjectMetaFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterObjectMetaFunc must not be empty", config)
	}

	r := &Resource{
		baseClusterConfig:        config.BaseClusterConfig,
		logger:                   config.Logger,
		tenant:                   config.Tenant,
		toClusterGuestConfigFunc: config.ToClusterGuestConfigFunc,
		toClusterObjectMetaFunc:  config.ToClusterObjectMetaFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ensureTillerInstalled(ctx context.Context, obj interface{}) error {
	objectMeta, err := r.toClusterObjectMetaFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// Tenant Tiller is not deleted by cluster-operator. Deleting tenant
	// cluster resources is handled by the provider operator
	// e.g. aws-operator.
	if key.IsDeleted(objectMeta) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not deleting tiller in tenant cluster")
		r.logger.LogCtx(ctx, "level", "debug", "message", "tiller in tenant cluster will be deleted with cluster deletion")

		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring tiller is installed")

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

	tenantHelmClient, err := r.tenant.NewHelmClient(ctx, clusterConfig.ClusterID, tenantAPIDomain)
	if err != nil {
		return microerror.Mask(err)
	}

	values := []string{
		"spec.template.spec.priorityClassName=giantswarm-critical",
		"spec.template.spec.tolerations[0].effect=NoSchedule",
		"spec.template.spec.tolerations[0].key=node-role.kubernetes.io/master",
		"spec.template.spec.tolerations[0].operator=Exists",
	}
	err = tenantHelmClient.EnsureTillerInstalledWithValues(ctx, values)
	if tenantcluster.IsTimeout(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "timeout fetching certificates")

		// A timeout error here means that the cluster-operator certificate
		// for the current guest cluster was not found. We can't continue
		// without a Helm client. We will retry during the next execution, when
		// the certificate might be available.
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)

		return nil
	} else if helmclient.IsTillerNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "Tiller installation failed")

		// Tiller installation can fail during guest cluster setup. We will
		// retry on next reconciliation loop.
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)

		return nil
	} else if tenant.IsAPINotAvailable(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "guest API not available")

		// We should not hammer guest API if it is not available, the guest
		// cluster might be initializing. We will retry on next reconciliation
		// loop.
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)

		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "ensured tiller is installed")

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
