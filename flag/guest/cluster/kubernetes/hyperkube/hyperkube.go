package hyperkube

import "github.com/giantswarm/kubernetesd/flag/service/cluster/kubernetes/hyperkube/docker"

// Hyperkube is a data structure to hold guest cluster Kubernetes Hyperkube
// image specific configuration flags.
type Hyperkube struct {
	Docker docker.Docker
}
