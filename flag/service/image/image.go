package image

import "github.com/giantswarm/cluster-operator/v3/flag/service/image/registry"

// Image is a data structure to hold container image specific configuration
// flags.
type Image struct {
	Registry registry.Registry
}
