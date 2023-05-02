package podcidr

import (
	"context"
	"strconv"
	"testing"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/operatorkit/v8/pkg/controller/context/cachekeycontext"

	"github.com/giantswarm/cluster-operator/v5/service/internal/unittest"
)

func Test_PodCIDR_Cache(t *testing.T) {
	testCases := []struct {
		name          string
		ctx           context.Context
		cidrBlock     string
		expectCaching bool
		expectCIDR    string
	}{
		{
			name:          "case 0",
			ctx:           cachekeycontext.NewContext(context.Background(), "1"),
			cidrBlock:     "pod-cidr",
			expectCaching: true,
			expectCIDR:    "pod-cidr",
		},
		// This is the case where we modify the AWSCluster CR in order to change the
		// pod CIDR value, while the operatorkit caching mechanism is disabled.
		{
			name:          "case 1",
			ctx:           context.Background(),
			cidrBlock:     "",
			expectCaching: false,
			expectCIDR:    "changed",
		},
		{
			name:          "case 2",
			ctx:           cachekeycontext.NewContext(context.Background(), "1"),
			cidrBlock:     "",
			expectCaching: true,
			expectCIDR:    "installation-cidr",
		},
		{
			name:          "case 3",
			ctx:           context.Background(),
			cidrBlock:     "pod-cidr",
			expectCaching: false,
			expectCIDR:    "changed",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var podCIDR1 string
			var podCIDR2 string

			var pc *PodCIDR
			{
				c := Config{
					K8sClient: unittest.FakeK8sClient(),

					InstallationCIDR: "installation-cidr",
					Provider:         "aws",
				}

				pc, err = New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			var cl infrastructurev1alpha3.AWSCluster
			{
				cl = unittest.DefaultCluster()
			}

			{
				cl.Spec.Provider.Pods.CIDRBlock = tc.cidrBlock
				err = pc.k8sClient.CtrlClient().Create(tc.ctx, &cl)
				if err != nil {
					t.Fatal(err)
				}
			}

			{
				podCIDR1, err = pc.PodCIDR(tc.ctx, &cl)
				if err != nil {
					t.Fatal(err)
				}
			}

			{
				cl.Spec.Provider.Pods.CIDRBlock = "changed"
				err = pc.k8sClient.CtrlClient().Update(tc.ctx, &cl)
				if err != nil {
					t.Fatal(err)
				}
			}

			{
				podCIDR2, err = pc.PodCIDR(tc.ctx, &cl)
				if err != nil {
					t.Fatal(err)
				}
			}

			if podCIDR2 != tc.expectCIDR {
				t.Fatalf("expected %#q to be equal to %#q", tc.expectCIDR, podCIDR2)
			}
			if tc.expectCaching {
				if podCIDR1 != podCIDR2 {
					t.Fatalf("expected %#q to be equal to %#q", podCIDR1, podCIDR2)
				}
			} else {
				if podCIDR1 == podCIDR2 {
					t.Fatalf("expected %#q to differ from %#q", podCIDR1, podCIDR2)
				}
			}
		})
	}
}
