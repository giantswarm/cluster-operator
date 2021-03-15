package cluster

import (
	"github.com/giantswarm/cluster-operator/v3/flag/guest/cluster/calico"
	"github.com/giantswarm/cluster-operator/v3/flag/guest/cluster/kubernetes"
)

// Cluster is a data structure to hold cluster specific configuration flags.
type Cluster struct {
	BaseDomain string
	Calico     calico.Calico
	Kubernetes kubernetes.Kubernetes
}
