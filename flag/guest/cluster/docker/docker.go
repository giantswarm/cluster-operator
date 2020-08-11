package docker

import "github.com/giantswarm/cluster-operator/v3/flag/guest/cluster/docker/daemon"

// Docker is a data structure to hold guest cluster Docker specific
// configuration flags.
type Docker struct {
	Daemon daemon.Daemon
}
