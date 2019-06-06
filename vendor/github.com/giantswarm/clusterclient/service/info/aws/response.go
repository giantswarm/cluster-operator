package aws

import (
	"github.com/giantswarm/clusterclient/service/info/aws/response"
)

// Response is the return value of the service action.
type Response struct {
	General response.General `json:"general"`
	Workers response.Workers `json:"workers"`
}
