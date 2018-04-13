package guestcluster

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Interface interface {
	// NewG8sClient returns a new generated clientset for a guest cluster.
	NewG8sClient(ctx context.Context, clusterID, apiDomain string) (versioned.Interface, error)
	// NewK8sClient returns a new Kubernetes clientset for a guest cluster.
	NewK8sClient(ctx context.Context, clusterID, apiDomain string) (kubernetes.Interface, error)
	// NewRestConfig returns a new REST config for a guest cluster.
	NewRestConfig(ctx context.Context, clusterID, apiDomain string) (*rest.Config, error)
}
