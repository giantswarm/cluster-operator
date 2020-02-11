package config

import "github.com/giantswarm/versionbundle"

// Patch is the cluster specific configuration.
type Patch struct {
	Name           string                 `json:"name"`
	Owner          string                 `json:"owner"`
	ReleaseVersion string                 `json:"release_version,omitempty"`
	Scaling        Scaling                `json:"scaling,omitempty"`
	VersionBundles []versionbundle.Bundle `json:"version_bundles,omitempty"`
	Workers        []Worker               `json:"workers,omitempty"`
}

// DefaultPatch provides a default patch by best effort.
func DefaultPatch() Patch {
	return Patch{
		Name:           "",
		Owner:          "",
		ReleaseVersion: "",
		Scaling:        Scaling{},
		VersionBundles: []versionbundle.Bundle{},
		Workers:        []Worker{},
	}
}
