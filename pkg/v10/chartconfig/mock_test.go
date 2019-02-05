package chartconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/helmclient"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
)

type tenantMock struct {
	fakeTenantG8sClient    versioned.Interface
	fakeTenantHelmClient   helmclient.Interface
	fakeTenantK8sClient    kubernetes.Interface
	fakeTenantK8sExtClient apiextensionsclient.Interface
}

func (g *tenantMock) NewG8sClient(ctx context.Context, clusterID, apiDomain string) (versioned.Interface, error) {
	return g.fakeTenantG8sClient, nil
}
func (g *tenantMock) NewHelmClient(ctx context.Context, clusterID, apiDomain string) (helmclient.Interface, error) {
	return g.fakeTenantHelmClient, nil
}
func (g *tenantMock) NewK8sClient(ctx context.Context, clusterID, apiDomain string) (kubernetes.Interface, error) {
	return g.fakeTenantK8sClient, nil
}
func (g *tenantMock) NewK8sExtClient(ctx context.Context, clusterID, apiDomain string) (apiextensionsclient.Interface, error) {
	return g.fakeTenantK8sExtClient, nil
}
