package searcher

import "github.com/giantswarm/versionbundle"

type Response struct {
	Active         bool                   `json:"active"`
	ReleaseVersion string                 `json:"release_version"`
	VersionBundles []versionbundle.Bundle `json:"version_bundles"`
}

func DefaultResponse() Response {
	return Response{
		Active:         false,
		ReleaseVersion: "",
		VersionBundles: []versionbundle.Bundle{},
	}
}
