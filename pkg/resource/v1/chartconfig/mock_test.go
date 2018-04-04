package chartconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
)

type guestClusterServiceMock struct {
	fakeGuestG8sClient versioned.Interface
}

func (g *guestClusterServiceMock) GetG8sClient(ctx context.Context, clusterID, apiDomain string) (versioned.Interface, error) {
	return g.fakeGuestG8sClient, nil
}
