package hyperkube

import "github.com/giantswarm/cluster-operator/v3/flag/guest/cluster/kubernetes/hyperkube/docker"

// Hyperkube is a data structure to hold guest cluster Kubernetes Hyperkube
// image specific configuration flags.
type Hyperkube struct {
	Docker docker.Docker
}
