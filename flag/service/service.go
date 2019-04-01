package service

import (
	"github.com/giantswarm/operatorkit/flag/service/kubernetes"

	"github.com/giantswarm/cluster-operator/flag/service/image"
	"github.com/giantswarm/cluster-operator/flag/service/kubeconfig"
)

// Service is an intermediate data structure for command line configuration flags.
type Service struct {
	Image      image.Image
	KubeConfig kubeconfig.KubeConfig
	Kubernetes kubernetes.Kubernetes
}
