package guestcluster

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
)

type Interface interface {
	GetG8sClient(ctx context.Context, clusterID, apiDomain string) (versioned.Interface, error)
}
