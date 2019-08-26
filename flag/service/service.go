package service

import (
	"github.com/giantswarm/operatorkit/flag/service/kubernetes"

	"github.com/giantswarm/cluster-operator/flag/service/clusterservice"
	"github.com/giantswarm/cluster-operator/flag/service/image"
	"github.com/giantswarm/cluster-operator/flag/service/kubeconfig"
	"github.com/giantswarm/cluster-operator/flag/service/provider"
)

// Service is an intermediate data structure for command line configuration flags.
type Service struct {
	ClusterService clusterservice.ClusterService
	Image          image.Image
	KubeConfig     kubeconfig.KubeConfig
	Kubernetes     kubernetes.Kubernetes
	Provider       provider.Provider
}
