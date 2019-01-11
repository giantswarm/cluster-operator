package configmap

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/helmclient"
	"k8s.io/client-go/kubernetes"
)

type tenantMock struct {
	fakeTenantG8sClient  versioned.Interface
	fakeTenantHelmClient helmclient.Interface
	fakeTenantK8sClient  kubernetes.Interface
}

func (t *tenantMock) NewG8sClient(ctx context.Context, clusterID, apiDomain string) (versioned.Interface, error) {
	return t.fakeTenantG8sClient, nil
}
func (t *tenantMock) NewHelmClient(ctx context.Context, clusterID, apiDomain string) (helmclient.Interface, error) {
	return t.fakeTenantHelmClient, nil
}
func (t *tenantMock) NewK8sClient(ctx context.Context, clusterID, apiDomain string) (kubernetes.Interface, error) {
	return t.fakeTenantK8sClient, nil
}
