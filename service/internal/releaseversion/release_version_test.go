package releaseversion

import (
	"context"
	"strconv"
	"testing"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha3"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/operatorkit/v5/pkg/controller/context/cachekeycontext"

	"github.com/giantswarm/cluster-operator/v3/service/internal/unittest"
)

func Test_Release_Cache(t *testing.T) {
	testCases := []struct {
		name             string
		ctx              context.Context
		appName          string
		releaseApp       ReleaseApp
		expectCaching    bool
		expectAppVersion string
	}{
		{
			name:    "case 0",
			ctx:     cachekeycontext.NewContext(context.Background(), "1"),
			appName: "cert-operator",
			releaseApp: ReleaseApp{
				Catalog: "default",
				Version: "1.2.1",
			},
			expectCaching:    true,
			expectAppVersion: "1.2.1",
		},
		// This is the case where we modify the Release CR in order to change the
		// app version value, while the operatorkit caching mechanism is disabled.
		{
			name:    "case 1",
			ctx:     context.Background(),
			appName: "cert-operator",
			releaseApp: ReleaseApp{
				Catalog: "default",
				Version: "1.2.1",
			},
			expectCaching:    false,
			expectAppVersion: "1.2.2",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var release1 map[string]ReleaseApp
			var release2 map[string]ReleaseApp

			var rv *ReleaseVersion
			{
				c := Config{
					K8sClient: unittest.FakeK8sClient(),
				}
				rv, err = New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			var release releasev1alpha1.Release
			{
				release = unittest.DefaultRelease()
			}

			var cl infrastructurev1alpha3.AWSCluster
			{
				cl = unittest.DefaultCluster()
			}

			{
				// Specify the version of the app we want for our tests.
				release.Spec.Apps[0].Version = tc.releaseApp.Version
				err = rv.k8sClient.CtrlClient().Create(tc.ctx, &release)
				if err != nil {
					t.Fatal(err)
				}
			}

			{
				release1, err = rv.Apps(tc.ctx, &cl)
				if err != nil {
					t.Fatal(err)
				}
			}

			{
				// Specify the updated version of the cert-operator we want for our tests.
				// The version should not change when caching is enabled.
				release.Spec.Apps[0].Version = "1.2.2"
				err = rv.k8sClient.CtrlClient().Update(tc.ctx, &release)
				if err != nil {
					t.Fatal(err)
				}
			}
			{
				release2, err = rv.Apps(tc.ctx, &cl)
				if err != nil {
					t.Fatal(err)
				}
			}

			if release2[tc.appName].Version != tc.expectAppVersion {
				t.Fatalf("expected %#q to be equal to %#q", release1[tc.appName], tc.expectAppVersion)
			}
			if tc.expectCaching {
				if release1[tc.appName] != release2[tc.appName] {
					t.Fatalf("expected %#q to be equal to %#q", release1[tc.appName], release2[tc.appName])
				}
			} else {
				if release1[tc.appName] == release2[tc.appName] {
					t.Fatalf("expected %#q to differ from %#q", release1[tc.appName], release1[tc.appName])
				}
			}

		})
	}
}
