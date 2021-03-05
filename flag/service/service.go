package service

import (
	"github.com/giantswarm/operatorkit/v4/pkg/flag/service/kubernetes"

	"github.com/giantswarm/cluster-operator/v3/flag/service/image"
	"github.com/giantswarm/cluster-operator/v3/flag/service/kubeconfig"
	"github.com/giantswarm/cluster-operator/v3/flag/service/release"
)

// Service is an intermediate data structure for command line configuration flags.
type Service struct {
	Image      image.Image
	KubeConfig kubeconfig.KubeConfig
	Kubernetes kubernetes.Kubernetes
	Release    release.Release
}
