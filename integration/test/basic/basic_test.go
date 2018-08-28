// +build k8srequired

package basic

import (
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
)

const (
	guestNamespace = "giantswarm"
	releaseName    = "chart-operator"
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

	err = waitForReleaseStatus(guestHelmClient, releaseName, "DEPLOYED")
	if err != nil {
		t.Fatalf("failed to get DEPLOYED status for release %#q", releaseName)
	}
}

// TestChartConfigChartsInstalled checks that the charts for any chartconfig
// CRs installed in the cluster have been deployed.
func TestChartConfigChartsInstalled(t *testing.T) {
	// These versions have no chartconfigs so we return early.
	clusterOperatorVersion := os.Getenv("CLOP_VERSION_BUNDLE_VERSION")
	if clusterOperatorVersion == "0.1.0" || clusterOperatorVersion == "0.2.0" {
		return
	}

	guestNamespace := "giantswarm"
	guestG8sClient := g.G8sClient()

	// Wait for chart configs as they may not have been created yet.
	err := waitForChartConfigs(guestG8sClient)
	if err != nil {
		t.Fatalf("could not get chartconfigs %v", err)
	}

	chartConfigList, err := guestG8sClient.CoreV1alpha1().ChartConfigs(guestNamespace).List(metav1.ListOptions{})
	if err != nil {
		t.Fatalf("could not get chartconfigs %v", err)
	}

	// At least 1 chartconfig is expected. Actual number depends on how many
	// components have been migrated.
	if len(chartConfigList.Items) == 0 {
		t.Fatalf("expected at least 1 chartconfigs: %d found", len(chartConfigList.Items))
	}
}

func TestChartConfigPatch(t *testing.T) {
	// These versions have no chartconfigs so we return early.
	clusterOperatorVersion := os.Getenv("CLOP_VERSION_BUNDLE_VERSION")
	if clusterOperatorVersion == "0.1.0" || clusterOperatorVersion == "0.2.0" {
		return
	}

	guestNamespace := "giantswarm"
	guestG8sClient := g.G8sClient()

	// Wait for chart configs as they may not have been created yet.
	err := waitForChartConfigs(guestG8sClient)
	if err != nil {
		t.Fatalf("could not get chartconfigs %v", err)
	}

	chartConfigList, err := guestG8sClient.CoreV1alpha1().ChartConfigs(guestNamespace).List(metav1.ListOptions{})
	if err != nil {
		t.Fatalf("could not get chartconfigs %v", err)
	}
	log.Printf("CHARTCONFIG_LIST: %v", chartConfigList)
	log.Printf("CHARTCONFIG_0: %s", chartConfigList.Items[0].Spec.Chart.Name)

	chartConfigName := chartConfigList.Items[0].Spec.Chart.Name
	chartConfig, err := guestG8sClient.CoreV1alpha1().ChartConfigs(guestNamespace).Get(chartConfigName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("could not get chartconfigs %v", err)
	}
	log.Printf("CHARTCONFIG:: %v", chartConfig)

	type patchStringValue struct {
		Op    string `json:"op"`
		Path  string `json:"path"`
		Value string `json:"value"`
	}
	payload := []patchStringValue{{
		Op:    "replace",
		Path:  "/spec/labels/giantswarm.io/managed-by",
		Value: "e2e-test",
	}}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("could not marshal json patch %v", err)
	}

	patchedChartConfig, err := guestG8sClient.CoreV1alpha1().ChartConfigs(guestNamespace).Patch(chartConfigName, types.JSONPatchType, payloadBytes)
	if err != nil {
		t.Fatalf("could not patch chartconfig %v", err)
	}
	log.Printf("CHARTCONFIG_PATCH: %v", patchedChartConfig)
}

func waitForChartConfigs(guestG8sClient versioned.Interface) error {
	operation := func() error {
		cc, err := guestG8sClient.CoreV1alpha1().ChartConfigs(guestNamespace).List(metav1.ListOptions{})
		if err != nil {
			return microerror.Mask(err)
		} else if len(cc.Items) == 0 {
			return microerror.Maskf(emptyChartConfigListError, "waiting for chart configs")
		}

		return nil
	}

	notify := func(err error, t time.Duration) {
		log.Printf("getting chart configs %s: %v", t, err)
	}

	b := backoff.NewExponential(10*time.Minute, framework.LongMaxInterval)
	err := backoff.RetryNotify(operation, b, notify)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func waitForReleaseStatus(guestHelmClient *helmclient.Client, release string, status string) error {
	operation := func() error {
		rc, err := guestHelmClient.GetReleaseContent(release)
		if err != nil {
			return microerror.Mask(err)
		}
		if rc.Status != status {
			return microerror.Maskf(releaseStatusNotMatchingError, "waiting for %q, current %q", status, rc.Status)
		}
		return nil
	}

	notify := func(err error, t time.Duration) {
		log.Printf("getting release status %s: %v", t, err)
	}

	b := backoff.NewExponential(20*time.Minute, framework.LongMaxInterval)
	err := backoff.RetryNotify(operation, b, notify)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}
