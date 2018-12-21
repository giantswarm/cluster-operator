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

func (a *apprMock) GetReleaseVersion(ctx context.Context, name, channel string) (string, error) {
	if a.defaultError {
		return "", fmt.Errorf("error getting default release")
	}

	return a.defaultReleaseVersion, nil
}

func (a *apprMock) PullChartTarball(ctx context.Context, name, channel string) (string, error) {
	return "", nil
}

func (a *apprMock) PullChartTarballFromRelease(ctx context.Context, name, release string) (string, error) {
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

func (h *helmMock) DeleteRelease(ctx context.Context, releaseName string, options ...helm.DeleteOption) error {
	if h.defaultError != nil {
		return h.defaultError
	}

	return nil
}

func (h *helmMock) GetReleaseContent(ctx context.Context, releaseName string) (*helmclient.ReleaseContent, error) {
	if h.defaultError != nil {
		return nil, h.defaultError
	}

	return h.defaultReleaseContent, nil
}

func (h *helmMock) GetReleaseHistory(ctx context.Context, releaseName string) (*helmclient.ReleaseHistory, error) {
	if h.defaultError != nil {
		return nil, h.defaultError
	}

	return h.defaultReleaseHistory, nil
}

func (h *helmMock) InstallReleaseFromTarball(ctx context.Context, path, ns string, options ...helm.InstallOption) error {
	return nil
}

func (h *helmMock) ListReleaseContents(ctx context.Context) ([]*helmclient.ReleaseContent, error) {
	return nil, nil
}

func (h *helmMock) PingTiller(ctx context.Context) error {
	return nil
}

func (h *helmMock) RunReleaseTest(ctx context.Context, releaseName string, options ...helm.ReleaseTestOption) error {
	return nil
}

func (h *helmMock) UpdateReleaseFromTarball(ctx context.Context, releaseName, path string, options ...helm.UpdateOption) error {
	return nil
}

func (h *helmMock) EnsureTillerInstalled(ctx context.Context) error {
	return nil
}
