package lister

import "github.com/giantswarm/clusterclient/service/cluster/lister/config"

// Request is the configuration for the service action.
type Request struct {
	Organization *config.Organization `json:"organization"`
}

// DefaultRequest provides a default request by best effort.
func DefaultRequest() Request {
	return Request{
		Organization: config.DefaultOrganization(),
	}
}
