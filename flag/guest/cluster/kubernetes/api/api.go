package api

// API is a data structure to hold guest cluster Kubernetes API specific
// configuration flags.
type API struct {
	AltNames       string
	ClusterIPRange string
	InsecurePort   string
	SecurePort     string
}
