package statuscondition

import (
	"context"
	"strconv"
	"testing"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclienttest"
	"github.com/giantswarm/micrologger/microloggertest"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	v1 "k8s.io/api/core/v1"

	"github.com/giantswarm/cluster-operator/v3/service/internal/recorder"
	"github.com/giantswarm/cluster-operator/v3/service/internal/releaseversion"
	tcunittest "github.com/giantswarm/cluster-operator/v3/service/internal/tenantclient/unittest"
	"github.com/giantswarm/cluster-operator/v3/service/internal/unittest"
)

type nodeConfig struct {
	replicas      int
	readyReplicas int
}

func TestComputeCreateClusterStatusConditions(t *testing.T) {
	testCases := []struct {
		name         string
		controlPlane nodeConfig
		nodePools    []nodeConfig
		release      v1alpha1.Release

		expectCondition string
	}{
		// This is the case where we simulating a cluster upgrade with condition `Updating` and we expect condition `Updated` to be set
		{
			name:            "case 0",
			release:         unittest.DefaultRelease(),
			controlPlane:    nodeConfig{1, 1},
			nodePools:       []nodeConfig{{1, 1}},
			expectCondition: "Updated",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			fakek8sclient := unittest.FakeK8sClient()
			ctx := context.Background()
			cluster := unittest.DefaultCluster()

			// The worker and master nodes are created
			var nodes []v1.Node
			var cps []infrastructurev1alpha2.G8sControlPlane
			var mds []apiv1alpha2.MachineDeployment
			{
				for _, nodePool := range tc.nodePools {
					md := unittest.DefaultMachineDeployment()
					md.Status.Replicas = int32(nodePool.replicas)
					md.Status.ReadyReplicas = int32(nodePool.readyReplicas)
					mds = append(mds, md)
					for i := 0; i < nodePool.readyReplicas; i++ {
						workerNode := unittest.NewWorkerNode()
						nodes = append(nodes, workerNode)
					}
				}
				cp := unittest.DefaultControlPlane()
				cp.Status.Replicas = int32(tc.controlPlane.replicas)
				cp.Status.ReadyReplicas = int32(tc.controlPlane.readyReplicas)
				cps = append(cps, cp)
				for i := 0; i < tc.controlPlane.readyReplicas; i++ {
					masterNode := unittest.NewMasterNode()
					nodes = append(nodes, masterNode)
				}
			}

			var err error
			var e recorder.Interface
			{
				c := recorder.Config{
					K8sClient: k8sclienttest.NewEmpty(),
				}
				e = recorder.New(c)
				if err != nil {
					t.Fatal(err)
				}
			}
			var rv *releaseversion.ReleaseVersion
			{
				c := releaseversion.Config{
					K8sClient: fakek8sclient,
				}
				rv, err = releaseversion.New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			r := Resource{
				event:                      e,
				k8sClient:                  fakek8sclient,
				logger:                     microloggertest.New(),
				releaseVersion:             rv,
				tenantClient:               tcunittest.FakeTenantClient(fakek8sclient),
				newCommonClusterObjectFunc: newCommonClusterObjectFunc("aws"),
				provider:                   "aws",
			}

			err = fakek8sclient.CtrlClient().Create(ctx, &tc.release)
			if err != nil {
				t.Fatal(err)
			}

			cl := apiv1alpha2.Cluster{}
			err = r.computeCreateClusterStatusConditions(ctx, cl, &cluster, nodes, cps, mds)
			if err != nil {
				t.Fatal(err)
			}
			status := cluster.GetCommonClusterStatus()

			if status.Conditions[0].Condition != tc.expectCondition {
				t.Fatalf("expected %#q to differ from %#q", tc.expectCondition, status.Conditions[0].Condition)
			}
		})
	}
}

func newCommonClusterObjectFunc(provider string) func() infrastructurev1alpha2.CommonClusterObject {
	// Deal with different providers in here once they reach Cluster API.
	return func() infrastructurev1alpha2.CommonClusterObject {
		return new(infrastructurev1alpha2.AWSCluster)
	}
}
