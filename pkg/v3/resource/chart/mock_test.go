package chart

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/helmclient"
	"k8s.io/client-go/kubernetes"
	"k8s.io/helm/pkg/helm"
)

type apprMock struct {
	defaultReleaseVersion string
	defaultError          bool
}

func (a *apprMock) GetReleaseVersion(name, channel string) (string, error) {
	if a.defaultError {
		return "", fmt.Errorf("error getting default release")
	}

	return a.defaultReleaseVersion, nil
}

func (a *apprMock) PullChartTarball(name, channel string) (string, error) {
	return "", nil
}

type guestMock struct {
	fakeGuestG8sClient  versioned.Interface
	fakeGuestHelmClient helmclient.Interface
	fakeGuestK8sClient  kubernetes.Interface
}

func (g *guestMock) NewG8sClient(ctx context.Context, clusterID, apiDomain string) (versioned.Interface, error) {
	return g.fakeGuestG8sClient, nil
}
func (g *guestMock) NewHelmClient(ctx context.Context, clusterID, apiDomain string) (helmclient.Interface, error) {
	return g.fakeGuestHelmClient, nil
}
func (g *guestMock) NewK8sClient(ctx context.Context, clusterID, apiDomain string) (kubernetes.Interface, error) {
	return g.fakeGuestK8sClient, nil
}

type helmMock struct {
	defaultReleaseContent *helmclient.ReleaseContent
	defaultReleaseHistory *helmclient.ReleaseHistory
	defaultError          error
}

func (h *helmMock) DeleteRelease(releaseName string, options ...helm.DeleteOption) error {
	if h.defaultError != nil {
		return h.defaultError
	}

	return nil
}

func (h *helmMock) GetReleaseContent(releaseName string) (*helmclient.ReleaseContent, error) {
	if h.defaultError != nil {
		return nil, h.defaultError
	}

	return h.defaultReleaseContent, nil
}

func (h *helmMock) GetReleaseHistory(releaseName string) (*helmclient.ReleaseHistory, error) {
	if h.defaultError != nil {
		return nil, h.defaultError
	}

	return h.defaultReleaseHistory, nil
}

func (h *helmMock) InstallFromTarball(path, ns string, options ...helm.InstallOption) error {
	return nil
}

func (h *helmMock) UpdateReleaseFromTarball(releaseName, path string, options ...helm.UpdateOption) error {
	return nil
}
