package nodecount

import (
	"context"
)

type Interface interface {
	// NodeCount is a map of key value pairs where the key is the control plane
	// ID. The map value is a structure holding node information for the
	// corresponding control plane.
	NodeCount(ctx context.Context, label string, obj interface{}) (map[string]Node, error)
}

// Node holds the node information for a control plane
type Node struct {
	Nodes int32
	Ready int32
}
