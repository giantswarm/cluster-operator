// +build k8srequired

package setup

import (
	"log"
	"os"
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/cluster-operator/integration/teardown"
)

func WrapTestMain(f *framework.Host, helmClient *helmclient.Client, m *testing.M) {
	var v int
	var err error

	err = f.CreateNamespace("giantswarm")
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	err = resources(f, helmClient)
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	if v == 0 {
		v = m.Run()
	}

	if os.Getenv("KEEP_RESOURCES") != "true" {
		// only do full teardown when not on CI
		if os.Getenv("CIRCLECI") != "true" {
			err := teardown.Teardown(f)
			if err != nil {
				log.Printf("%#v\n", err)
				v = 1
			}
			// TODO there should be error handling for the framework teardown.
			f.Teardown()
		}
	}

	os.Exit(v)
}

func resources(f *framework.Host, helmClient *helmclient.Client) error {
	const clusterOperatorValues = `Installation:
  V1:
    Guest:
      Kubernetes:
        API:
          ClusterIPRange: 10.0.0.0/16
    Auth:
      Vault:
        Certificate:
          TTL: 3000h
    Secret:
      Registry:
        PullSecret:
          DockerConfigJSON: "{\"auths\":{\"quay.io\":{\"auth\":\"${REGISTRY_PULL_SECRET}\"}}}"
`

	err := f.InstallOperator("cluster-operator", "awsclusterconfig", clusterOperatorValues, "@1.0.0-${CIRCLE_SHA1}")

	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
