package nodecount

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/operatorkit/v2/pkg/controller/context/cachekeycontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tcunittest "github.com/giantswarm/cluster-operator/v3/service/internal/tenantclient/unittest"
	"github.com/giantswarm/cluster-operator/v3/service/internal/unittest"
)

func Test_NodeCount_Cache(t *testing.T) {
	testCases := []struct {
		name            string
		ctx             context.Context
		nodeCount       int32
		expectCaching   bool
		expectNodeCount int32
	}{
		{
			name:            "case 0",
			ctx:             cachekeycontext.NewContext(context.Background(), "1"),
			nodeCount:       1,
			expectCaching:   true,
			expectNodeCount: 1,
		},
		// This is the case where we modify the AWSCluster CR in order to change the
		// baseDomain value, while the operatorkit caching mechanism is disabled.
		{
			name:            "case 1",
			ctx:             context.Background(),
			nodeCount:       1,
			expectCaching:   false,
			expectNodeCount: 2,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var masterNodes1 map[string]Node
			var masterNodes2 map[string]Node
			var controlPlaneValue = "xyz123"
			var controlPlaneKey = "giantswarm.io/control-plane"

			var nc *NodeCount
			{
				fakeK8sClient := unittest.FakeK8sClient()
				c := Config{
					K8sClient:    fakeK8sClient,
					TenantClient: tcunittest.FakeTenantClient(fakeK8sClient),
				}

				nc, err = New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			{
				nodes := unittest.DefaultNodes()
				for _, node := range nodes.Items {
					nodeLabels := node.GetLabels()
					if _, ok := nodeLabels[controlPlaneKey]; ok {
						nodeLabels[controlPlaneKey] = controlPlaneValue
					}
					node.SetLabels(nodeLabels)
					_, err := nc.k8sClient.K8sClient().CoreV1().Nodes().Create(tc.ctx, node.DeepCopy(), metav1.CreateOptions{})
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			{
				cl := unittest.DefaultCluster()
				masterNodes1, err = nc.MasterCount(tc.ctx, &cl)
				if err != nil {
					t.Fatal(err)
				}
			}

			newNode := unittest.NewMasterNode()
			{
				nodeLabels := newNode.GetLabels()
				if _, ok := nodeLabels[controlPlaneKey]; ok {
					nodeLabels[controlPlaneKey] = controlPlaneValue
				}
				newNode.ObjectMeta.Name = "ip-10-0-5-50.eu-central-1.compute.internal"
				newNode.SetLabels(nodeLabels)
				_, err = nc.k8sClient.K8sClient().CoreV1().Nodes().Create(tc.ctx, &newNode, metav1.CreateOptions{})
				if err != nil {
					t.Fatal(err)
				}
			}

			{
				cl := unittest.DefaultCluster()
				masterNodes2, err = nc.MasterCount(tc.ctx, &cl)
				if err != nil {
					t.Fatal(err)
				}
			}

			if masterNodes2[controlPlaneValue].Nodes != tc.expectNodeCount {
				t.Fatalf("expected %v to be equal to %v", tc.expectNodeCount, masterNodes2[controlPlaneValue].Nodes)
			}
			if tc.expectCaching {
				if masterNodes1[controlPlaneValue].Nodes != masterNodes2[controlPlaneValue].Nodes {
					t.Fatalf("expected %v to be equal to %v", masterNodes1[controlPlaneValue].Nodes, masterNodes2[controlPlaneValue].Nodes)
				}
			} else {
				if masterNodes1[controlPlaneValue].Nodes == masterNodes2[controlPlaneValue].Nodes {
					t.Fatalf("expected %v to differ from %v", masterNodes1[controlPlaneValue].Nodes, masterNodes2[controlPlaneValue].Nodes)
				}
			}
		})
	}
}
