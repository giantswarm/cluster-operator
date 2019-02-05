package namespace

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/helmclient"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
)

type guestMock struct {
	fakeGuestG8sClient    versioned.Interface
	fakeGuestHelmClient   helmclient.Interface
	fakeGuestK8sClient    kubernetes.Interface
	fakeGuestK8sExtClient apiextensionsclient.Interface
}

func (g *guestMock) NewG8sClient(ctx context.Context, clusterID, apiDomain string) (versioned.Interface, error) {
	return g.fakeGuestG8sClient, nil
}
func (g *guestMock) NewHelmClient(ctx context.Context, clusterID, apiDomain string) (helmclient.Interface, error) {
	return g.fakeGuestHelmClient, nil
}
func (g *guestMock) NewK8sClient(ctx context.Context, clusterID, apiDomain string) (kubernetes.Interface, error) {
	return g.fakeGuestK8sClient, nil
}
func (g *guestMock) NewK8sExtClient(ctx context.Context, clusterID, apiDomain string) (apiextensionsclient.Interface, error) {
	return g.fakeGuestK8sExtClient, nil
}
