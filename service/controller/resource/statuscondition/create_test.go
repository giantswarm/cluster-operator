package statuscondition

import (
	"context"
	"strconv"
	"testing"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclienttest"
	"github.com/giantswarm/micrologger/microloggertest"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	v1 "k8s.io/api/core/v1"

	"github.com/giantswarm/cluster-operator/v3/service/internal/recorder"
	"github.com/giantswarm/cluster-operator/v3/service/internal/releaseversion"
	tcunittest "github.com/giantswarm/cluster-operator/v3/service/internal/tenantclient/unittest"
	"github.com/giantswarm/cluster-operator/v3/service/internal/unittest"
)

func TestComputeCreateClusterStatusConditions(t *testing.T) {
	testCases := []struct {
		name          string
		cluster       infrastructurev1alpha2.AWSCluster
		ctx           context.Context
		fakek8sclient k8sclient.Interface
		release       v1alpha1.Release

		expectCondition string
	}{
		// This is the case where we simulating a cluster upgrade with condition `Updating` and we expect condition `Updated` to be set
		{
			name:          "case 0",
			cluster:       unittest.DefaultCluster(),
			ctx:           context.Background(),
			fakek8sclient: unittest.FakeK8sClient(),
			release:       unittest.DefaultRelease(),

			expectCondition: "Updated",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {

			workerNode := unittest.NewWorkerNode()
			masterNode := unittest.NewMasterNode()
			nodes := []v1.Node{masterNode, workerNode}

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
					K8sClient: tc.fakek8sclient,
				}
				rv, err = releaseversion.New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			r := Resource{
				event:                      e,
				k8sClient:                  tc.fakek8sclient,
				logger:                     microloggertest.New(),
				releaseVersion:             rv,
				tenantClient:               tcunittest.FakeTenantClient(tc.fakek8sclient),
				newCommonClusterObjectFunc: newCommonClusterObjectFunc("aws"),
				provider:                   "aws",
			}

			err = tc.fakek8sclient.CtrlClient().Create(tc.ctx, &tc.release)
			if err != nil {
				t.Fatal(err)
			}

			cps := []infrastructurev1alpha2.G8sControlPlane{unittest.DefaultControlPlane()}
			mds := []apiv1alpha2.MachineDeployment{unittest.DefaultMachineDeployment()}

			err = r.computeCreateClusterStatusConditions(tc.ctx, &tc.cluster, nodes, cps, mds)
			if err != nil {
				t.Fatal(err)
			}
			status := tc.cluster.GetCommonClusterStatus()

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
