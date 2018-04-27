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
	h          *framework.Host
	g          *framework.Guest
	helmClient *helmclient.Client
	apprClient *apprclient.Client
)

// TestMain allows us to have common setup and teardown steps that are run
// once for all the tests https://golang.org/pkg/testing/#hdr-Main.
func TestMain(m *testing.M) {
	var err error

	h, err = framework.NewHost(framework.HostConfig{})
	if err != nil {
		panic(err.Error())
	}

	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		panic(err.Error())
	}
	g, err = framework.NewGuest(framework.GuestConfig{Logger: logger})
	if err != nil {
		panic(err.Error())
	}

	ch := helmclient.Config{
		Logger:     logger,
		K8sClient:  h.K8sClient(),
		RestConfig: h.RestConfig(),
	}
	helmClient, err = helmclient.New(ch)
	if err != nil {
		panic(err.Error())
	}

	fs := afero.NewOsFs()
	ca := apprclient.Config{
		Fs:     fs,
		Logger: logger,

		Address:      "https://quay.io",
		Organization: "giantswarm",
	}

	apprClient, err = apprclient.New(ca)
	if err != nil {
		panic(err.Error())
	}

	setup.WrapTestMain(g, h, helmClient, apprClient, m)
}
