package creator

import (
	"github.com/giantswarm/clusterclient/service/keypair/creator/config"
)

// Request is the configuration for the service action.
type Request struct {
	Cluster *config.Cluster `json:"cluster"`
	KeyPair *config.KeyPair `json:"key_pair"`
}

// DefaultRequest provides a default request by best effort.
func DefaultRequest() Request {
	return Request{
		Cluster: config.DefaultCluster(),
		KeyPair: config.DefaultKeyPair(),
	}
}
