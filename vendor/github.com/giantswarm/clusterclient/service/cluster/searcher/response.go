package searcher

import (
	"time"

	"github.com/giantswarm/versionbundle"

	"github.com/giantswarm/clusterclient/service/cluster/searcher/response"
	"github.com/giantswarm/clusterclient/service/cluster/searcher/response/aws"
	"github.com/giantswarm/clusterclient/service/cluster/searcher/response/kvm"
)

// Response is the return value of the service action.
type Response struct {
	APIEndpoint       string                 `json:"api_endpoint"`
	AvailabilityZones []string               `json:"availability_zones,omitempty"`
	AWS               aws.Cluster            `json:"aws,omitempty"`
	CreateDate        time.Time              `json:"create_date"`
	ID                string                 `json:"id"`
	KVM               kvm.Cluster            `json:"kvm,omitempty"`
	Masters           []response.Master      `json:"masters,omitempty"`
	Name              string                 `json:"name,omitempty"`
	Owner             string                 `json:"owner,omitempty"`
	ReleaseVersion    string                 `json:"release_version,omitempty"`
	Scaling           response.Scaling       `json:"scaling,omitempty"`
	VersionBundles    []versionbundle.Bundle `json:"version_bundles,omitempty"`
	Workers           []response.Worker      `json:"workers,omitempty"`
	CredentialID      string                 `json:"credential_id"`
}

// DefaultResponse provides a default response by best effort.
func DefaultResponse() *Response {
	return &Response{
		APIEndpoint:       "",
		AvailabilityZones: []string{},
		AWS:               aws.DefaultCluster(),
		CreateDate:        time.Time{},
		ID:                "",
		KVM:               kvm.Cluster{},
		Masters:           []response.Master{},
		Name:              "",
		Owner:             "",
		ReleaseVersion:    "",
		Scaling:           response.Scaling{},
		VersionBundles:    []versionbundle.Bundle{},
		Workers:           []response.Worker{},
		CredentialID:      "",
	}
}
