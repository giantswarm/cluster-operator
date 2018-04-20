// +build k8srequired

package basic

import (
	"testing"

	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"

	"github.com/giantswarm/cluster-operator/integration/setup"
)

var (
	f          *framework.Host
	helmClient *helmclient.Client
	apprClient *apprclient.Client
)

// TestMain allows us to have common setup and teardown steps that are run
// once for all the tests https://golang.org/pkg/testing/#hdr-Main.
func TestMain(m *testing.M) {
	var err error

	f, err = framework.NewHost(framework.HostConfig{})
	if err != nil {
		panic(err.Error())
	}

	l, err := micrologger.New(micrologger.Config{})
	if err != nil {
		panic(err.Error())
	}

	ch := helmclient.Config{
		Logger:     l,
		K8sClient:  f.K8sClient(),
		RestConfig: f.RestConfig(),
	}
	helmClient, err = helmclient.New(ch)
	if err != nil {
		panic(err.Error())
	}

	fs := afero.NewOsFs()
	ca := apprclient.Config{
		Fs:     fs,
		Logger: l,

		Address:      "https://quay.io",
		Organization: "giantswarm",
	}

	apprClient, err = apprclient.New(ca)
	if err != nil {
		panic(err.Error())
	}

	setup.WrapTestMain(f, helmClient, m)
}
