// +build k8srequired

package basic

import (
	"testing"

	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"

	"github.com/giantswarm/cluster-operator/integration/env"
	"github.com/giantswarm/cluster-operator/integration/setup"
)

var (
	g          *framework.Guest
	h          *framework.Host
	l          micrologger.Logger
	apprClient *apprclient.Client
	helmClient *helmclient.Client
)

func init() {
	var err error

	{
		c := micrologger.Config{}

		l, err = micrologger.New(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c := framework.GuestConfig{
			Logger: l,

			ClusterID:    env.ClusterID(),
			CommonDomain: env.CommonDomain(),
		}

		g, err = framework.NewGuest(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c := framework.HostConfig{
			Logger: l,

			ClusterID: env.ClusterID(),
		}

		h, err = framework.NewHost(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c := helmclient.Config{
			Logger:     l,
			K8sClient:  h.K8sClient(),
			RestConfig: h.RestConfig(),
		}

		helmClient, err = helmclient.New(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c := apprclient.Config{
			Fs:     afero.NewOsFs(),
			Logger: l,

			Address:      "https://quay.io",
			Organization: "giantswarm",
		}

		apprClient, err = apprclient.New(c)
		if err != nil {
			panic(err.Error())
		}
	}
}

// TestMain allows us to have common setup and teardown steps that are run
// once for all the tests https://golang.org/pkg/testing/#hdr-Main.
func TestMain(m *testing.M) {
	setup.WrapTestMain(g, h, helmClient, apprClient, m)
}
