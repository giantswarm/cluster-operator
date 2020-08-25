package statuscondition

import (
	"context"
	"testing"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/micrologger/microloggertest"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/cluster-operator/v3/service/internal/releaseversion"
	tcunittest "github.com/giantswarm/cluster-operator/v3/service/internal/tenantclient/unittest"
	"github.com/giantswarm/cluster-operator/v3/service/internal/unittest"
	v1 "k8s.io/api/core/v1"
)

func TestComputeCreateClusterStatusConditions(t *testing.T) {
	fakeK8sClient := unittest.FakeK8sClient()
	workerNode := unittest.NewWorkerNode()
	masterNode := unittest.NewMasterNode()
	nodes := []v1.Node{masterNode, workerNode}

	var rv *releaseversion.ReleaseVersion
	var err error
	{
		c := releaseversion.Config{
			K8sClient: fakeK8sClient,
		}
		rv, err = releaseversion.New(c)
		if err != nil {
			t.Fatal(err)
		}
	}

	r := Resource{
		k8sClient:                  fakeK8sClient,
		logger:                     microloggertest.New(),
		releaseVersion:             rv,
		tenantClient:               tcunittest.FakeTenantClient(fakeK8sClient),
		newCommonClusterObjectFunc: newCommonClusterObjectFunc("aws"),
		provider:                   "aws",
	}
	ctx := context.TODO()

	// create a new release
	release := unittest.DefaultRelease()
	fakeK8sClient.CtrlClient().Create(ctx, release.DeepCopy())

	cluster := unittest.DefaultCluster()
	cps := []infrastructurev1alpha2.G8sControlPlane{unittest.DefaultControlPlane()}
	mds := []apiv1alpha2.MachineDeployment{unittest.DefaultMachineDeployment()}

	err = r.computeCreateClusterStatusConditions(ctx, &cluster, nodes, cps, mds)
	if err != nil {
		t.Fatal(err)
	}
	status := cluster.GetCommonClusterStatus()

	if !(status.Conditions[0].Condition == "Updated") {
		t.Fatal("First condition has to be 'Updated', we expect status condition to be set")
	}

}

func newCommonClusterObjectFunc(provider string) func() infrastructurev1alpha2.CommonClusterObject {
	// Deal with different providers in here once they reach Cluster API.
	return func() infrastructurev1alpha2.CommonClusterObject {
		return new(infrastructurev1alpha2.AWSCluster)
	}
}
