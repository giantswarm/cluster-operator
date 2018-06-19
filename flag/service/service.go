package service

import (
	"github.com/giantswarm/cluster-operator/flag/service/kubernetes"
	"github.com/giantswarm/cluster-operator/flag/service/storage"
)

// Service is an intermediate data structure for command line configuration flags.
type Service struct {
	Kubernetes kubernetes.Kubernetes
	Storage    storage.Storage
}
