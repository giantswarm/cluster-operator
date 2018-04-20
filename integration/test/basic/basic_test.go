package basic

import (
	"io/ioutil"
	"os"
	"os/exec"
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

	err = setupDNS()
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

	err = teardownDNS()
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

func setupDNS() error {
	// /etc/hosts is in a file system mounted as readonly, we will bind mount the modified file.
	content, err := ioutil.ReadFile("/etc/hosts")
	if err != nil {
		return microerror.Maskf(err, "could not read hosts file %v")
	}
	err = ioutil.WriteFile(tmpHostsFile, content, 0644)
	if err != nil {
		return microerror.Maskf(err, "could not write tmp hosts file %v")
	}

	f, err := os.OpenFile(tmpHostsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return microerror.Maskf(err, "could not open hosts file %v")
	}
	defer f.Close()

	dnsEntry := os.ExpandEnv("\n127.0.0.1     api.${CLUSTER_NAME}.aws.giantswarm.io")
	if _, err := f.Write([]byte(dnsEntry)); err != nil {
		return microerror.Maskf(err, "could not append new entry to tmp hosts file %v")
	}

	cmd := exec.Command("mount", "--bind", tmpHostsFile, "/etc/hosts")
	err = cmd.Run()
	if err != nil {
		return microerror.Maskf(err, "could not bind mount tmp hosts file")
	}

	return nil
}

func setupIPTables() error {

	return nil
}

func setupCertificates() error {

	return nil
}

func teardownDNS() error {
	cmd := exec.Command("umount", "/etc/hosts")
	err := cmd.Run()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func teardownIPTables() error {

	return nil
}

func teardownCertificates() error {

	return nil
}
