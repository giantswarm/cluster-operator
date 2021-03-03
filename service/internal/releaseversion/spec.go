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
	Apps(ctx context.Context, obj interface{}) (map[string]ReleaseApp, error)
	// ComponentVersion provides the version of each component in a release.
	ComponentVersion(ctx context.Context, obj interface{}) (map[string]ReleaseComponent, error)
}

type ReleaseApp struct {
	// Catalog of the app.
	Catalog string `json:"catalog"`
	// Version of the app.
	Version string `json:"version"`
}

type ReleaseComponent struct {
	// Catalog of the component.
	Catalog string `json:"catalog"`
	// Version of the component in the catalog.
	Reference string `json:"reference"`
	// Version of the component.
	Version string `json:"version"`
}
