package basic

import (
	"testing"

	"github.com/giantswarm/microerror"
	"k8s.io/helm/pkg/helm"
)

const (
	tmpHostsFile = "/home/e2e-harness/hosts.new"
)

func TestChartOperatorBootstrap(t *testing.T) {
	err := setUp()
	if err != nil {
		t.Fatalf("could not setup test: %v", err)
	}
	// defer tearDown()

}

func installResource(name string) error {
	tarball, err := apprClient.PullChartTarball(name+"-chart", "stable")
	if err != nil {
		return microerror.Mask(err)
	}
	err = helmClient.InstallFromTarball(tarball, "default",
		helm.ReleaseName(name),
		helm.ValueOverrides([]byte("{}")),
		helm.InstallWait(true))
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func removeResource(name string) error {
	err := helmClient.DeleteRelease(name,
		helm.DeletePurge(true),
	)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func setUp() error {
	err := installResource("cluster-operator-resource")
	if err != nil {
		return microerror.Mask(err)
	}

	err = setupIPTables()
	if err != nil {
		return microerror.Mask(err)
	}

	err = setupCertificates()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func tearDown() error {
	err := removeResource("cluster-operator-resource")
	if err != nil {
		return microerror.Mask(err)
	}

	err = teardownIPTables()
	if err != nil {
		return microerror.Mask(err)
	}

	err = teardownCertificates()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func setupIPTables() error {

	return nil
}

func setupCertificates() error {

	return nil
}

func teardownIPTables() error {

	return nil
}

func teardownCertificates() error {

	return nil
}
