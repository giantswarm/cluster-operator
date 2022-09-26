package basedomain

import (
	"context"
	"strconv"
	"testing"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/cachekeycontext"

	"github.com/giantswarm/cluster-operator/v5/service/internal/unittest"
)

func Test_BaseDomain_Cache(t *testing.T) {
	testCases := []struct {
		name             string
		ctx              context.Context
		baseDomain       string
		expectCaching    bool
		expectBaseDomain string
	}{
		{
			name:             "case 0",
			ctx:              cachekeycontext.NewContext(context.Background(), "1"),
			baseDomain:       "domain.company.com",
			expectCaching:    true,
			expectBaseDomain: "domain.company.com",
		},
		// This is the case where we modify the AWSCluster CR in order to change the
		// baseDomain value, while the operatorkit caching mechanism is disabled.
		{
			name:             "case 1",
			ctx:              context.Background(),
			baseDomain:       "olddomain.company.com",
			expectCaching:    false,
			expectBaseDomain: "newdomain.company.com",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var baseDomain1 string
			var baseDomain2 string

			var bd *BaseDomain
			{
				c := Config{
					K8sClient: unittest.FakeK8sClient(),
					Provider:  "aws",
				}

				bd, err = New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			var cl infrastructurev1alpha3.AWSCluster
			{
				cl = unittest.DefaultCluster()
			}

			{
				cl.Spec.Cluster.DNS.Domain = tc.baseDomain
				err = bd.k8sClient.CtrlClient().Create(tc.ctx, &cl)
				if err != nil {
					t.Fatal(err)
				}
			}

			{
				baseDomain1, err = bd.BaseDomain(tc.ctx, &cl)
				if err != nil {
					t.Fatal(err)
				}
			}

			{
				cl.Spec.Cluster.DNS.Domain = "newdomain.company.com"
				err = bd.k8sClient.CtrlClient().Update(tc.ctx, &cl)
				if err != nil {
					t.Fatal(err)
				}
			}

			{
				baseDomain2, err = bd.BaseDomain(tc.ctx, &cl)
				if err != nil {
					t.Fatal(err)
				}
			}

			if baseDomain2 != tc.expectBaseDomain {
				t.Fatalf("expected %#q to be equal to %#q", tc.expectBaseDomain, baseDomain2)
			}
			if tc.expectCaching {
				if baseDomain1 != baseDomain2 {
					t.Fatalf("expected %#q to be equal to %#q", baseDomain1, baseDomain2)
				}
			} else {
				if baseDomain1 == baseDomain2 {
					t.Fatalf("expected %#q to differ from %#q", baseDomain1, baseDomain2)
				}
			}
		})
	}
}
