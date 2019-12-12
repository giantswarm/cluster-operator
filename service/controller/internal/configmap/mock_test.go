package configmap

import (
	"context"

	"k8s.io/client-go/rest"
)

type tenantMock struct {
	fakeTenantRestConfig *rest.Config
}

func (t *tenantMock) NewRestConfig(ctx context.Context, clusterID, apiDomain string) (*rest.Config, error) {
	return t.fakeTenantRestConfig, nil
}
