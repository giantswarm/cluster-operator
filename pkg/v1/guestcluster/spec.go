package guestcluster

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"k8s.io/client-go/kubernetes"
)

type Interface interface {
	// NewG8sClient returns a new generated clientset for a guest cluster.
	NewG8sClient(ctx context.Context, clusterID, apiDomain string) (versioned.Interface, error)
	// NewK8sClient returns a new Kubernetes clientset for a guest cluster.
	NewK8sClient(ctx context.Context, clusterID, apiDomain string) (kubernetes.Interface, error)
}
