package flag

import (
	"github.com/giantswarm/microkit/flag"

	"github.com/giantswarm/cluster-operator/v5/flag/guest"
	"github.com/giantswarm/cluster-operator/v5/flag/service"
)

// Flag provides data structure for service command line flags.
type Flag struct {
	Guest   guest.Guest
	Service service.Service
}

// New constructs fills new Flag structure with given command line flags.
func New() *Flag {
	f := &Flag{}
	flag.Init(f)

	return f
}
