package releaseversion

import (
	"context"
)

const (
	// AppOperator defines the name of the app operator in a release.
	AppOperator = "app-operator"
	// CertOperator defines the name of the certificate operator in a release.
	CertOperator = "cert-operator"
)

type Interface interface {
	// AppVersion provides the version of each app in a release.
	AppVersion(ctx context.Context, obj interface{}) (map[string]string, error)
	// ComponentVersion provides the version of each component in a release.
	ComponentVersion(ctx context.Context, obj interface{}) (map[string]string, error)
}
