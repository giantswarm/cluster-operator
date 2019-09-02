package controllercontext

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/helmclient"
	"k8s.io/client-go/kubernetes"
)

type ContextClient struct {
	TenantCluster ContextClientTenantCluster
}

type ContextClientTenantCluster struct {
	G8s  versioned.Interface
	Helm helmclient.Interface
	K8s  kubernetes.Interface
}
