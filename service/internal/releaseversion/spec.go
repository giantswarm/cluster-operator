package releaseversion

import (
	"context"
)

type Interface interface {
	// ReleaseVersioner TODO
	ReleaseVersioner(ctx context.Context, obj interface{}) (map[string]string, error)
}
