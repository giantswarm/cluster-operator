package lister

import "github.com/giantswarm/clusterclient/service/keypair/lister/config"

// Request is the configuration for the service action.
type Request struct {
	Cluster *config.Cluster `json:"cluster"`
}

// DefaultRequest provides a default request by best effort.
func DefaultRequest() Request {
	return Request{
		Cluster: config.DefaultCluster(),
	}
}
