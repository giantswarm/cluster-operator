package creator

import (
	"github.com/giantswarm/clusterclient/service/cluster/creator/response"
)

type responseError struct {
	Code  string
	Error string
}

// Response is the return value of the service action.
type Response struct {
	Cluster response.Cluster `json:"cluster"`
}

// DefaultResponse provides a default response object by best effort.
func DefaultResponse() *Response {
	return &Response{
		Cluster: response.DefaultCluster(),
	}
}
