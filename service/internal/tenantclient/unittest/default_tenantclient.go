package unittest

import (
	"context"

	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"

	"github.com/giantswarm/cluster-operator/service/internal/tenantclient"
)

type fakeTenantClient struct {
	k8sClient k8sclient.Interface
}

func FakeTenantClient(k8sclient k8sclient.Interface) tenantclient.Interface {
	var tenantClient tenantclient.Interface
	{
		tenantClient = &fakeTenantClient{
			k8sClient: k8sclient,
		}
	}

	return tenantClient
}
func (f *fakeTenantClient) K8sClient(ctx context.Context, obj interface{}) (k8sclient.Interface, error) {
	return f.k8sClient, nil
}
