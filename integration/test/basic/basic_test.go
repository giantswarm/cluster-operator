// +build k8srequired

package basic

import (
	"testing"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/micrologger"
)

const (
	releaseName = "chart-operator"
)

func TestChartOperatorBootstrap(t *testing.T) {
	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("could not create logger %v", err)
	}

	ch := helmclient.Config{
		Logger:          logger,
		K8sClient:       g.K8sClient(),
		RestConfig:      g.RestConfig(),
		TillerNamespace: "giantswarm",
	}
	guestHelmClient, err := helmclient.New(ch)
	if err != nil {
		t.Fatalf("could not create guest helm client %v", err)
	}

	releaseContent, err := guestHelmClient.GetReleaseContent(releaseName)
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
