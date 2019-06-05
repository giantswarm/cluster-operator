package response

// Cluster is an object containing cluster specific information.
type Cluster struct {
	ID string `json:"id"`
}

// DefaultCluster provides a default cluster object by best effort.
func DefaultCluster() Cluster {
	return Cluster{
		ID: "",
	}
}
