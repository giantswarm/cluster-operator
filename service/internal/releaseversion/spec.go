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
	// AppVersion TODO
	AppVersion(ctx context.Context, obj interface{}) (map[string]string, error)
	// ComponentVersion TODO
	ComponentVersion(ctx context.Context, obj interface{}) (map[string]string, error)
}
