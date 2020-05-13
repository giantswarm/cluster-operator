package basedomain

import (
	"context"
)

type Interface interface {
	// BaseDomain TODO
	BaseDomain(ctx context.Context, obj interface{}) (string, error)
}
