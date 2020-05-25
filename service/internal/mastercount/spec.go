package mastercount

import (
	"context"
)

type Interface interface {
	// MasterCount is a map of key value pairs where the key is the control plane
	// ID. The map value is a structure holding node information for the
	// corresponding control plane.
	MasterCount(ctx context.Context, obj interface{}) (map[string]Master, error)
}

// Master holds the Master Node information for a control plane
type Master struct {
	Nodes int32
	Ready int32
}
