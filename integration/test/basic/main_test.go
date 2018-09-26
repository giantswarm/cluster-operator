// +build k8srequired

package basic

import (
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2e-harness/pkg/framework/resource"
	e2eclient "github.com/giantswarm/e2eclients/aws"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/cluster-operator/integration/env"
	"github.com/giantswarm/cluster-operator/integration/setup"
)

var (
	c          *e2eclient.Client
	g          *framework.Guest
	h          *framework.Host
	helmClient *helmclient.Client
	l          micrologger.Logger
	r          *resource.Resource
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

			ClusterID:  env.ClusterID(),
			VaultToken: env.VaultToken(),
		}

		h, err = framework.NewHost(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c := helmclient.Config{
			Logger:          l,
			K8sClient:       h.K8sClient(),
			RestConfig:      h.RestConfig(),
			TillerNamespace: "giantswarm",
		}
		helmClient, err = helmclient.New(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c := resource.Config{
			Logger:     l,
			HelmClient: helmClient,
			Namespace:  "giantswarm",
		}
		r, err = resource.New(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c, err = e2eclient.NewClient()
		if err != nil {
			panic(err.Error())
		}
	}
}

// TestMain allows us to have common setup and teardown steps that are run
// once for all the tests https://golang.org/pkg/testing/#hdr-Main.
func TestMain(m *testing.M) {
	{
		c := setup.Config{
			AWSClient: c,
			Guest:     g,
			Host:      h,
			Logger:    l,
			Resource:  r,
		}

		setup.Setup(m, c)
	}
}
