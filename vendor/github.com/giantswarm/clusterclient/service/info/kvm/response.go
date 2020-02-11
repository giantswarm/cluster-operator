package kvm

import (
	"github.com/giantswarm/clusterclient/service/info/kvm/response"
)

// Response is the return value of the service action.
type Response struct {
	General response.General `json:"general"`
	Workers response.Workers `json:"workers"`
}

// DefaultResponse provides a default response by best effort.
func DefaultResponse() Response {
	return Response{
		General: response.DefaultGeneral(),
		Workers: response.DefaultWorkers(),
	}
}
