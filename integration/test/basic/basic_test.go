// +build k8srequired

package basic

import (
	"os"
	"testing"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	releaseName = "chart-operator"
)

func TestChartOperatorBootstrap(t *testing.T) {
	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("could not create logger %v", err)
	}

	tillerNamespace := "giantswarm"

	// This version bundle uses kube-system because it doesn't have the
	// namespace resource that creates the giantswarm namespace. All future
	// version will use the giantswarm namespace.
	if os.Getenv("CLOP_VERSION_BUNDLE_VERSION") == "0.2.0" {
		tillerNamespace = "kube-system"
	}

	ch := helmclient.Config{
		Logger:          logger,
		K8sClient:       g.K8sClient(),
		RestConfig:      g.RestConfig(),
		TillerNamespace: tillerNamespace,
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

func TestChartConfigChartsInstalled(t *testing.T) {
	guestNamespace := "giantswarm"
	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("could not create logger %v", err)
	}

	guestG8sClient := g.G8sClient()
	chartConfigList, err := guestG8sClient.CoreV1alpha1().ChartConfigs(guestNamespace).List(metav1.ListOptions{})
	if err != nil {
		t.Fatalf("could not get chartconfigs %v", err)
	}

	if len(chartConfigList.Items) > 0 {
		ch := helmclient.Config{
			Logger:          logger,
			K8sClient:       g.K8sClient(),
			RestConfig:      g.RestConfig(),
			TillerNamespace: guestNamespace,
		}
		guestHelmClient, err := helmclient.New(ch)
		if err != nil {
			t.Fatalf("could not create guest helm client %v", err)
		}

		for _, chart := range chartConfigList.Items {
			releaseName := chart.Spec.Chart.Release
			releaseContent, err := guestHelmClient.GetReleaseContent(releaseName)
			if err != nil {
				t.Fatalf("could not get release content for release %q %v", releaseName, err)
			}

			expectedStatus := "DEPLOYED"
			actualStatus := releaseContent.Status
			if expectedStatus != actualStatus {
				t.Fatalf("bad release status for %q, want %q, got %q", releaseName, expectedStatus, actualStatus)
			}
		}
	}
}
