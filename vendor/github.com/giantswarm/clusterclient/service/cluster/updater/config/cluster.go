package config

// Cluster is the cluster specific configuration.
type Cluster struct {
	ID    string `json:"id"`
	Patch Patch  `json:"patch"`
}

// DefaultCluster provides a default cluster by best effort.
func DefaultCluster() Cluster {
	return Cluster{
		ID:    "",
		Patch: DefaultPatch(),
	}
}
