package kubernetes

import (
	"github.com/giantswarm/cluster-operator/v3/flag/guest/cluster/kubernetes/api"
	"github.com/giantswarm/cluster-operator/v3/flag/guest/cluster/kubernetes/hyperkube"
	"github.com/giantswarm/cluster-operator/v3/flag/guest/cluster/kubernetes/ingresscontroller"
	"github.com/giantswarm/cluster-operator/v3/flag/guest/cluster/kubernetes/kubelet"
	"github.com/giantswarm/cluster-operator/v3/flag/guest/cluster/kubernetes/networksetup"
	"github.com/giantswarm/cluster-operator/v3/flag/guest/cluster/kubernetes/ssh"
)

// Kubernetes is a data structure to hold guest cluster Kubernetes specific
// configuration flags.
type Kubernetes struct {
	API               api.API
	ClusterDomain     string
	Hyperkube         hyperkube.Hyperkube
	IngressController ingresscontroller.IngressController
	Kubelet           kubelet.Kubelet
	NetworkSetup      networksetup.NetworkSetup
	SSH               ssh.SSH
}
