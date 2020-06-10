package tenantclient

import (
	"context"

	client "github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
)

type Interface interface {
	// K8sClient returns client interface of the corresponding cluster object
	K8sClient(ctx context.Context, obj interface{}) (client.Interface, error)
}
