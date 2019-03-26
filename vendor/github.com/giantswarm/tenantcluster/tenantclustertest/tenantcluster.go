package tenantclustertest

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/tenantcluster"
	"k8s.io/client-go/kubernetes"
)

type Config struct {
	G8sClient       versioned.Interface
	G8sClientError  error
	HelmClient      helmclient.Interface
	HelmClientError error
	K8sClient       kubernetes.Interface
	K8sClientError  error
}

type TenantCluster struct {
	g8sClient       versioned.Interface
	g8sClientError  error
	helmClient      helmclient.Interface
	helmClientError error
	k8sClient       kubernetes.Interface
	k8sClientError  error
}

func New(config Config) tenantcluster.Interface {
	t := &TenantCluster{
		g8sClient:       config.G8sClient,
		g8sClientError:  config.G8sClientError,
		helmClient:      config.HelmClient,
		helmClientError: config.HelmClientError,
		k8sClient:       config.K8sClient,
		k8sClientError:  config.K8sClientError,
	}

	return t
}

func (t *TenantCluster) NewG8sClient(ctx context.Context, clusterID, apiDomain string) (versioned.Interface, error) {
	if t.g8sClientError != nil {
		return nil, t.g8sClientError
	}

	return t.g8sClient, nil
}

func (t *TenantCluster) NewHelmClient(ctx context.Context, clusterID, apiDomain string) (helmclient.Interface, error) {
	if t.helmClientError != nil {
		return nil, t.helmClientError
	}

	return t.helmClient, nil
}

func (t *TenantCluster) NewK8sClient(ctx context.Context, clusterID, apiDomain string) (kubernetes.Interface, error) {
	if t.k8sClientError != nil {
		return nil, t.k8sClientError
	}

	return t.k8sClient, nil
}
