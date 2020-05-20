package releaseversion

import (
	"context"
)

type Interface interface {
	// AppVersion TODO
	AppVersion(ctx context.Context, obj interface{}) (map[string]string, error)
	// ComponentVersion TODO
	ComponentVersion(ctx context.Context, obj interface{}) (map[string]string, error)
}
