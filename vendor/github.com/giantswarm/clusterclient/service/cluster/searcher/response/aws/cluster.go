package aws

// Cluster configures AWS-specific cluster settings.
type Cluster struct {
	ResourceTags map[string]string `json:"resource_tags"`
}

// DefaultCluster provides a default Cluster.
func DefaultCluster() Cluster {
	return Cluster{
		ResourceTags: map[string]string{},
	}
}
