package nodecount

import (
	"context"
)

type Interface interface {
	// MasterCount is a map of key value pairs where the key is the control plane
	// ID. The map value is a structure holding node information for the
	// corresponding control plane.
	MasterCount(ctx context.Context, obj interface{}) (map[string]Node, error)
	// WorkerCount is a map of key value pairs where the key is the machine deployment
	// ID. The map value is a structure holding node information for the
	// corresponding node pools.
	WorkerCount(ctx context.Context, obj interface{}) (map[string]Node, error)
}

// Node holds the node information for a control plane or a machine deployment
type Node struct {
	Nodes int32
	Ready int32
}
