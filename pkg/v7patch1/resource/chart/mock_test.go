package chart

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/helmclient"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/helm/pkg/helm"
)

type apprMock struct {
	defaultReleaseVersion string
	defaultError          bool
}

func (a *apprMock) DeleteRelease(ctx context.Context, name, channel string) error {
	return nil
}

func (a *apprMock) GetReleaseVersion(ctx context.Context, name, channel string) (string, error) {
	if a.defaultError {
		return "", fmt.Errorf("error getting default release")
	}

	return a.defaultReleaseVersion, nil
}

func (a *apprMock) PromoteChart(ctx context.Context, name, release, channel string) error {
	return nil
}

func (a *apprMock) PullChartTarball(ctx context.Context, name, channel string) (string, error) {
	return "", nil
}

func (a *apprMock) PullChartTarballFromRelease(ctx context.Context, name, release string) (string, error) {
	return "", nil
}

func (a *apprMock) PushChartTarball(ctx context.Context, name, release, tarballPath string) error {
	return nil
}

type tenantMock struct {
	fakeTenantG8sClient    versioned.Interface
	fakeTenantHelmClient   helmclient.Interface
	fakeTenantK8sClient    kubernetes.Interface
	fakeTenantK8sExtClient apiextensionsclient.Interface
}

func (g *tenantMock) NewG8sClient(ctx context.Context, clusterID, apiDomain string) (versioned.Interface, error) {
	return g.fakeTenantG8sClient, nil
}
func (g *tenantMock) NewHelmClient(ctx context.Context, clusterID, apiDomain string) (helmclient.Interface, error) {
	return g.fakeTenantHelmClient, nil
}
func (g *tenantMock) NewK8sClient(ctx context.Context, clusterID, apiDomain string) (kubernetes.Interface, error) {
	return g.fakeTenantK8sClient, nil
}
func (g *tenantMock) NewK8sExtClient(ctx context.Context, clusterID, apiDomain string) (apiextensionsclient.Interface, error) {
	return g.fakeTenantK8sExtClient, nil
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

func (h *helmMock) EnsureTillerInstalled(ctx context.Context) error {
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

func (h *helmMock) LoadChart(ctx context.Context, chartPath string) (helmclient.Chart, error) {
	return helmclient.Chart{}, nil
}

func (h *helmMock) PingTiller(ctx context.Context) error {
	return nil
}

func (h *helmMock) PullChartTarball(ctx context.Context, tarballURL string) (string, error) {
	return "", nil
}

func (h *helmMock) RunReleaseTest(ctx context.Context, releaseName string, options ...helm.ReleaseTestOption) error {
	return nil
}

func (h *helmMock) UpdateReleaseFromTarball(ctx context.Context, releaseName, path string, options ...helm.UpdateOption) error {
	return nil
}
