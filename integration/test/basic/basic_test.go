// +build k8srequired

package basic

import (
	"testing"
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

func setUp() error {

	return nil
}

func tearDown() error {

	return nil
}
