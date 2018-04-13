package chartconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type guestMock struct {
	fakeGuestG8sClient  versioned.Interface
	fakeGuestK8sClient  kubernetes.Interface
	fakeGuestRestConfig *rest.Config
}

func (g *guestMock) NewG8sClient(ctx context.Context, clusterID, apiDomain string) (versioned.Interface, error) {
	return g.fakeGuestG8sClient, nil
}
func (g *guestMock) NewK8sClient(ctx context.Context, clusterID, apiDomain string) (kubernetes.Interface, error) {
	return g.fakeGuestK8sClient, nil
}

func (g *guestMock) NewRestConfig(ctx context.Context, clusterID, apiDomain string) (*rest.Config, error) {
	return g.fakeGuestRestConfig, nil
}
