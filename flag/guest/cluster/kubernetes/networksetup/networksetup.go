package networksetup

import "github.com/giantswarm/kubernetesd/flag/service/cluster/kubernetes/networksetup/docker"

// NetworkSetup is a data structure to hold guest cluster network setup
// configuration flags.
type NetworkSetup struct {
	Docker docker.Docker
}
