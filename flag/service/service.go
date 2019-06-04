package service

import (
	"github.com/giantswarm/operatorkit/flag/service/kubernetes"

	"github.com/giantswarm/cluster-operator/flag/service/cluster"
	"github.com/giantswarm/cluster-operator/flag/service/image"
	"github.com/giantswarm/cluster-operator/flag/service/kubeconfig"
)

// Service is an intermediate data structure for command line configuration flags.
type Service struct {
	Cluster    cluster.Cluster
	Image      image.Image
	KubeConfig kubeconfig.KubeConfig
	Kubernetes kubernetes.Kubernetes
}
