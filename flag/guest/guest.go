package guest

import "github.com/giantswarm/cluster-operator/v5/flag/guest/cluster"

// Guest is a data structure to hold guest cluster specific configuration
// flags.
type Guest struct {
	Cluster cluster.Cluster
}
