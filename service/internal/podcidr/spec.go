package podcidr

import (
	"context"
)

type Interface interface {
	// PodCIDR provides the pod CIDR to be used for Tenant Clusters depending on
	// the installation and AWSCluster CR configuration. The CR value is prefered
	// over the default value in the installation.
	PodCIDR(ctx context.Context, obj interface{}) (string, error)
}
