package service

import (
	"github.com/giantswarm/operatorkit/flag/service/kubernetes"

	"github.com/giantswarm/cluster-operator/flag/service/image"
)

// Service is an intermediate data structure for command line configuration flags.
type Service struct {
	Image      image.Image
	Kubernetes kubernetes.Kubernetes
}
