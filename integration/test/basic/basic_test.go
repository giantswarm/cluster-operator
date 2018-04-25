// +build k8srequired

package basic

import (
	"testing"
)

const (
	releaseName = "chart-operator"
)

func TestChartOperatorBootstrap(t *testing.T) {
	releaseContent, err := helmClient.GetReleaseContent(releaseName)
	if err != nil {
		t.Fatalf("could not get release content %v", err)
	}

	expectedName := releaseName
	actualName := releaseContent.Name
	if expectedName != actualName {
		t.Fatalf("bad release name, want %q, got %q", expectedName, actualName)
	}

	expectedStatus := "DEPLOYED"
	actualStatus := releaseContent.Status
	if expectedStatus != actualStatus {
		t.Fatalf("bad release status, want %q, got %q", expectedStatus, actualStatus)
	}
}
