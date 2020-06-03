package client

import (
	"context"

	client "github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
)

type Interface interface {
	// K8sClient TODO
	K8sClient(ctx context.Context, obj interface{}) (client.Interface, error)
}
