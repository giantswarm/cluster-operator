package creator

import (
	"github.com/giantswarm/clusterclient/service/cluster/creator/request"
	"github.com/giantswarm/clusterclient/service/cluster/creator/request/aws"
	"github.com/giantswarm/versionbundle"
)

// Request is the configuration for the service action.
type Request struct {
	AvailabilityZones int                    `json:"availability_zones,omitempty"`
	AWS               aws.Cluster            `json:"aws,omitempty"`
	ID                string                 `json:"id,omitempty"`
	Masters           []request.Master       `json:"masters,omitempty"`
	Name              string                 `json:"name,omitempty"`
	Owner             string                 `json:"owner,omitempty"`
	ReleaseVersion    string                 `json:"release_version,omitempty"`
	Scaling           request.Scaling        `json:"scaling,omitempty"`
	VersionBundles    []versionbundle.Bundle `json:"version_bundles,omitempty"`
	Workers           []request.Worker       `json:"workers,omitempty"`
}
