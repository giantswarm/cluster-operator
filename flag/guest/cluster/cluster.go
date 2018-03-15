package cluster

import (
	"github.com/giantswarm/cluster-operator/flag/guest/cluster/calico"
	"github.com/giantswarm/cluster-operator/flag/guest/cluster/docker"
	"github.com/giantswarm/cluster-operator/flag/guest/cluster/etcd"
	"github.com/giantswarm/cluster-operator/flag/guest/cluster/kubernetes"
	"github.com/giantswarm/cluster-operator/flag/guest/cluster/provider"
	"github.com/giantswarm/cluster-operator/flag/guest/cluster/vault"
)

// Cluster is a data structure to hold cluster specific configuration flags.
type Cluster struct {
	Calico     calico.Calico
	Docker     docker.Docker
	Etcd       etcd.Etcd
	Kubernetes kubernetes.Kubernetes
	Provider   provider.Provider
	Vault      vault.Vault
}
