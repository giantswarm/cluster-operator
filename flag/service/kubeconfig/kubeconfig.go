package kubeconfig

import (
	"github.com/giantswarm/cluster-operator/flag/service/kubeconfig/resource"
)

// KubeConfig is a data structure to hold kubeconfig specific configuration flags.
type KubeConfig struct {
	Resource resource.Resource
}
