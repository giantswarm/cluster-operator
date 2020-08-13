package ingresscontroller

import "github.com/giantswarm/cluster-operator/v3/flag/guest/cluster/kubernetes/ingresscontroller/docker"

// IngressController is a data structure to hold guest cluster ingress
// controller specific configuration flags.
type IngressController struct {
	Docker docker.Docker
}
