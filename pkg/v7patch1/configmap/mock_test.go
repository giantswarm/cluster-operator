package configmap

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/helmclient"
	"k8s.io/client-go/kubernetes"
	"k8s.io/helm/pkg/helm"
)

type tenantMock struct {
	fakeTenantG8sClient  versioned.Interface
	fakeTenantHelmClient helmclient.Interface
	fakeTenantK8sClient  kubernetes.Interface
}

func (t *tenantMock) NewG8sClient(ctx context.Context, clusterID, apiDomain string) (versioned.Interface, error) {
	return t.fakeTenantG8sClient, nil
}
func (t *tenantMock) NewHelmClient(ctx context.Context, clusterID, apiDomain string) (helmclient.Interface, error) {
	return t.fakeTenantHelmClient, nil
}
func (t *tenantMock) NewK8sClient(ctx context.Context, clusterID, apiDomain string) (kubernetes.Interface, error) {
	return t.fakeTenantK8sClient, nil
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

func (h *helmMock) EnsureTillerInstalled() error {
	return nil
}

func (h *helmMock) EnsureTillerInstalledWithValues(ctx context.Context) error {
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

func (h *helmMock) InstallReleaseFromTarball(path, ns string, options ...helm.InstallOption) error {
	return nil
}

func (h *helmMock) ListReleaseContents(ctx context.Context) ([]*helmclient.ReleaseContent, error) {
	return nil, nil
}

func (h *helmMock) PingTiller() error {
	return nil
}

func (h *helmMock) RunReleaseTest(releaseName string, options ...helm.ReleaseTestOption) error {
	return nil
}

func (h *helmMock) UpdateReleaseFromTarball(releaseName, path string, options ...helm.UpdateOption) error {
	return nil
}
