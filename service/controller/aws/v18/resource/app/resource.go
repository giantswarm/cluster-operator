package app

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "appv18"
)

type StateGetter struct {
	// Dependencies.
	k8sClient kubernetes.Interface
	g8sClient versioned.Interface
	logger    micrologger.Logger

	projectName string
}

func (s *StateGetter) Name() string {
	return Name
}
