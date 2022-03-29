package kubeconfig

import (
	"github.com/giantswarm/cluster-operator/v4/flag/service/kubeconfig/resource"
)

// KubeConfig is a data structure to hold kubeconfig specific configuration flags.
type KubeConfig struct {
	Secret resource.Secret
}
