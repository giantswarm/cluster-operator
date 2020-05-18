package basedomain

import (
	"context"
)

type Interface interface {
	// BaseDomain provides the base domain from all tenant clusters.
	// It is used in all component certificates.
	BaseDomain(ctx context.Context, obj interface{}) (string, error)
}
