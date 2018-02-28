package etcd

// Etcd is a data structure to hold guest cluster Etcd specific configuration
// flags.
type Etcd struct {
	AltNames string
	Port     string
	Prefix   string
}
