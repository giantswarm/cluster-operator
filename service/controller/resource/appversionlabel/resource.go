package appversionlabel

import (
	"fmt"
	"strings"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Name = "appversionlabel"
)

type Config struct {
	G8sClient            versioned.Interface
	GetClusterConfigFunc func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
	Logger               micrologger.Logger
}

type Resource struct {
	g8sClient            versioned.Interface
	getClusterConfigFunc func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
	logger               micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		g8sClient: config.G8sClient,
		logger:    config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) getComponentVersion(releaseVersion, component string) (string, error) {
	release, err := r.g8sClient.ReleaseV1alpha1().Releases().Get(releaseVersion, metav1.GetOptions{})
	if err != nil {
		return "", microerror.Mask(err)
	}

	for _, c := range release.Spec.Components {
		if c.Name == component {
			return c.Version, nil
		}
	}

	return "", microerror.Maskf(notFoundError, fmt.Sprintf("can't find the release version %#q", releaseVersion))
}

func replaceToEscape(from string) string {
	return strings.Replace(from, "/", "~1", -1)
}
