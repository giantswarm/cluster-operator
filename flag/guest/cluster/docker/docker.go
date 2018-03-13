package docker

import "github.com/giantswarm/cluster-operator/flag/guest/cluster/docker/daemon"

// Docker is a data structure to hold guest cluster Docker specific
// configuration flags.
type Docker struct {
	Daemon daemon.Daemon
}
