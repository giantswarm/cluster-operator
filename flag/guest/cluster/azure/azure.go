package azure

import "github.com/giantswarm/cluster-operator/flag/guest/cluster/azure/hostcluster"

// Azure is an intermediate data structure for command line configuration flags.
type Azure struct {
	HostCluster hostcluster.HostCluster
}
