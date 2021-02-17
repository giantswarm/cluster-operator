package statuscondition

import (
	"context"
	"strconv"
	"testing"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclienttest"
	"github.com/giantswarm/micrologger/microloggertest"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/cluster-operator/v3/service/internal/recorder"
	"github.com/giantswarm/cluster-operator/v3/service/internal/unittest"
)

func TestComputeClusterStatusConditions(t *testing.T) {
	testCases := []struct {
		name string

		nodesReady      bool
		conditions      []infrastructurev1alpha2.CommonClusterStatusCondition
		versions        []infrastructurev1alpha2.CommonClusterStatusVersion
		operatorVersion string

		expectCondition string
		expectVersion   string
	}{
		// The cluster is creating
		{
			name: "case 0",

			nodesReady:      false,
			conditions:      []infrastructurev1alpha2.CommonClusterStatusCondition{},
			versions:        []infrastructurev1alpha2.CommonClusterStatusVersion{},
			operatorVersion: "8.7.5",

			expectCondition: "Creating",
			expectVersion:   "",
		},
		// The cluster is still creating - some nodes are not ready yet
		{
			name: "case 1",

			nodesReady: false,
			conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
				unittest.GetCreatingCondition(90),
			},
			versions:        []infrastructurev1alpha2.CommonClusterStatusVersion{},
			operatorVersion: "8.7.5",

			expectCondition: "Creating",
			expectVersion:   "",
		},
		// The cluster is created
		{
			name: "case 2",

			nodesReady: true,
			conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
				unittest.GetCreatingCondition(90),
			},
			versions:        []infrastructurev1alpha2.CommonClusterStatusVersion{},
			operatorVersion: "8.7.5",

			expectCondition: "Created",
			expectVersion:   "8.7.5",
		},
		// We add a nodepool
		{
			name: "case 3",

			nodesReady: false,
			conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
				unittest.GetCreatedCondition(60),
				unittest.GetCreatingCondition(90),
			},
			versions: []infrastructurev1alpha2.CommonClusterStatusVersion{
				unittest.GetVersion(60, "8.7.5"),
			},
			operatorVersion: "8.7.5",

			expectCondition: "Created",
			expectVersion:   "8.7.5",
		},
		// The cluster is upgrading
		{
			name: "case 4",

			nodesReady: false,
			conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
				unittest.GetCreatedCondition(60),
				unittest.GetCreatingCondition(90),
			},
			versions: []infrastructurev1alpha2.CommonClusterStatusVersion{
				unittest.GetVersion(60, "8.7.5"),
			},
			operatorVersion: "8.7.6",

			expectCondition: "Updating",
			expectVersion:   "8.7.5",
		},
		// Some nodes are not ready yet
		{
			name: "case 5",

			nodesReady: false,
			conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
				unittest.GetUpdatingCondition(15),
				unittest.GetCreatedCondition(60),
				unittest.GetCreatingCondition(90),
			},
			versions: []infrastructurev1alpha2.CommonClusterStatusVersion{
				unittest.GetVersion(60, "8.7.5"),
			},
			operatorVersion: "8.7.6",

			expectCondition: "Updating",
			expectVersion:   "8.7.5",
		},
		// This is the case where we simulating a cluster upgrade with condition `Updating` and we expect condition `Updated` to be set
		{
			name: "case 6",

			nodesReady: true,
			conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
				unittest.GetUpdatingCondition(15),
				unittest.GetCreatedCondition(60),
				unittest.GetCreatingCondition(90),
			},
			versions: []infrastructurev1alpha2.CommonClusterStatusVersion{
				unittest.GetVersion(60, "8.7.5"),
			},
			operatorVersion: "8.7.6",

			expectCondition: "Updated",
			expectVersion:   "8.7.6",
		},
		// We go HA masters
		{
			name: "case 7",

			nodesReady: false,
			conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
				unittest.GetUpdatedCondition(5),
				unittest.GetUpdatingCondition(15),
				unittest.GetCreatedCondition(60),
				unittest.GetCreatingCondition(90),
			},
			versions: []infrastructurev1alpha2.CommonClusterStatusVersion{
				unittest.GetVersion(5, "8.7.6"),
				unittest.GetVersion(60, "8.7.5"),
			},
			operatorVersion: "8.7.6",

			expectCondition: "Updated",
			expectVersion:   "8.7.6",
		},
		// We upgrade again
		{
			name: "case 8",

			nodesReady: false,
			conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
				unittest.GetUpdatedCondition(5),
				unittest.GetUpdatingCondition(15),
				unittest.GetCreatedCondition(60),
				unittest.GetCreatingCondition(90),
			},
			versions: []infrastructurev1alpha2.CommonClusterStatusVersion{
				unittest.GetVersion(5, "8.7.6"),
				unittest.GetVersion(60, "8.7.5"),
			},
			operatorVersion: "8.7.7",

			expectCondition: "Updating",
			expectVersion:   "8.7.6",
		},
		// We want to roll back
		{
			name: "case 9",

			nodesReady: false,
			conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
				unittest.GetUpdatedCondition(5),
				unittest.GetUpdatingCondition(15),
				unittest.GetCreatedCondition(60),
				unittest.GetCreatingCondition(90),
			},
			versions: []infrastructurev1alpha2.CommonClusterStatusVersion{
				unittest.GetVersion(5, "8.7.6"),
				unittest.GetVersion(60, "8.7.5"),
			},
			operatorVersion: "8.7.5",

			expectCondition: "Updating",
			expectVersion:   "8.7.6",
		},
		// We rolled back
		// TODO: when we roll back, the version does currently not appear in the history again.
		// this is because of how the function withVersion() in apiextensions works
		{
			name: "case 10",

			nodesReady: true,
			conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
				unittest.GetUpdatingCondition(5),
				unittest.GetUpdatedCondition(15),
				unittest.GetUpdatingCondition(20),
				unittest.GetCreatedCondition(60),
				unittest.GetCreatingCondition(90),
			},
			versions: []infrastructurev1alpha2.CommonClusterStatusVersion{
				unittest.GetVersion(15, "8.7.6"),
				unittest.GetVersion(60, "8.7.5"),
			},
			operatorVersion: "8.7.5",

			expectCondition: "Updated",
			expectVersion:   "8.7.6",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			fakek8sclient := unittest.FakeK8sClient()
			ctx := context.Background()

			// The cluster is created
			var cluster infrastructurev1alpha2.AWSCluster
			var cl apiv1alpha2.Cluster
			{
				cluster = unittest.DefaultCluster()
				cluster.Status.Cluster.Conditions = tc.conditions
				cluster.Status.Cluster.Versions = tc.versions
				cl = apiv1alpha2.Cluster{}
			}

			var e recorder.Interface
			{
				c := recorder.Config{
					K8sClient: k8sclienttest.NewEmpty(),
				}
				e = recorder.New(c)
			}

			r := Resource{
				event:     e,
				k8sClient: fakek8sclient,
				logger:    microloggertest.New(),
			}

			err := r.writeClusterStatusConditions(ctx, cl, &cluster, tc.nodesReady, tc.operatorVersion)
			if err != nil {
				t.Fatal(err)
			}

			// Check results
			status := cluster.GetCommonClusterStatus()

			if status.LatestCondition() != tc.expectCondition {
				t.Fatalf("expected %#q to differ from %#q", tc.expectCondition, status.LatestCondition())
			}
			if status.LatestVersion() != tc.expectVersion {
				t.Fatalf("expected %#q to differ from %#q", tc.expectVersion, status.LatestVersion())
			}
		})
	}
}

func TestComputeCreatingCondition(t *testing.T) {
	testCases := []struct {
		name string

		status         infrastructurev1alpha2.CommonClusterStatus
		expectedResult bool
	}{
		// There are no previous versions or conditions in the status
		{
			name: "case 0",

			status: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{},
				Versions:   []infrastructurev1alpha2.CommonClusterStatusVersion{},
			},
			expectedResult: true,
		},
		// There are previous versions in the status
		{
			name: "case 1",

			status: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{},
				Versions: []infrastructurev1alpha2.CommonClusterStatusVersion{
					unittest.GetVersion(5, "8.7.6"),
				},
			},
			expectedResult: false,
		},
		// There are previous conditions in the status
		{
			name: "case 2",

			status: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
					unittest.GetCreatingCondition(90),
				},
				Versions: []infrastructurev1alpha2.CommonClusterStatusVersion{},
			},
			expectedResult: false,
		},
		// There are previous conditions and versions in the status
		{
			name: "case 3",

			status: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
					unittest.GetCreatingCondition(90),
				},
				Versions: []infrastructurev1alpha2.CommonClusterStatusVersion{
					unittest.GetVersion(60, "8.7.6"),
				},
			},
			expectedResult: false,
		},
		// There are previous conditions and versions in the status
		{
			name: "case 4",

			status: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
					unittest.GetUpdatedCondition(10),
					unittest.GetUpdatingCondition(30),
					unittest.GetCreatedCondition(60),
					unittest.GetCreatingCondition(90),
				},
				Versions: []infrastructurev1alpha2.CommonClusterStatusVersion{
					unittest.GetVersion(60, "8.7.6"),
					unittest.GetVersion(60, "8.7.5"),
				},
			},
			expectedResult: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result := computeCreatingCondition(tc.status)
			if result != tc.expectedResult {
				t.Fatalf("expected %v, got %v", tc.expectedResult, result)
			}
		})
	}
}

func TestComputeCreatedCondition(t *testing.T) {
	testCases := []struct {
		name string

		status     infrastructurev1alpha2.CommonClusterStatus
		nodesReady bool

		expectedResult bool
	}{
		// the cluster is creating and nodes are ready
		{
			name: "case 0",

			status: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
					unittest.GetCreatingCondition(90),
				},
			},
			nodesReady: true,

			expectedResult: true,
		},
		// the cluster is creating but nodes are not ready
		{
			name: "case 1",

			status: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
					unittest.GetCreatingCondition(90),
				},
			},
			nodesReady: false,

			expectedResult: false,
		},
		// The cluster already has a created condition
		{
			name: "case 2",

			status: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
					unittest.GetCreatedCondition(60),
					unittest.GetCreatingCondition(90),
				},
			},
			nodesReady: true,

			expectedResult: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result := computeCreatedCondition(tc.status, tc.nodesReady)
			if result != tc.expectedResult {
				t.Fatalf("expected %v, got %v", tc.expectedResult, result)
			}
		})
	}
}

func TestComputeUpdatingCondition(t *testing.T) {
	testCases := []struct {
		name string

		status         infrastructurev1alpha2.CommonClusterStatus
		desiredVersion string

		expectedResult bool
	}{
		// the cluster is created and not in updating state but the desired version differs from the current version
		{
			name: "case 0",

			status: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
					unittest.GetCreatedCondition(60),
					unittest.GetCreatingCondition(90),
				},
				Versions: []infrastructurev1alpha2.CommonClusterStatusVersion{
					unittest.GetVersion(60, "8.7.5"),
				},
			},
			desiredVersion: "8.7.6",

			expectedResult: true,
		},
		// the cluster does not have a created condition
		{
			name: "case 1",

			status: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
					unittest.GetCreatingCondition(90),
				},
				Versions: []infrastructurev1alpha2.CommonClusterStatusVersion{
					unittest.GetVersion(60, "8.7.5"),
				},
			},
			desiredVersion: "8.7.6",

			expectedResult: false,
		},
		// the cluster is already updating
		{
			name: "case 2",

			status: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
					unittest.GetUpdatingCondition(30),
					unittest.GetCreatedCondition(60),
					unittest.GetCreatingCondition(90),
				},
				Versions: []infrastructurev1alpha2.CommonClusterStatusVersion{
					unittest.GetVersion(60, "8.7.5"),
				},
			},
			desiredVersion: "8.7.6",

			expectedResult: false,
		},
		// the version is already changed
		{
			name: "case 3",

			status: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
					unittest.GetCreatedCondition(60),
					unittest.GetCreatingCondition(90),
				},
				Versions: []infrastructurev1alpha2.CommonClusterStatusVersion{
					unittest.GetVersion(60, "8.7.6"),
				},
			},
			desiredVersion: "8.7.6",

			expectedResult: false,
		},
		// the version is already present and cluster is already updating
		{
			name: "case 4",

			status: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
					unittest.GetUpdatingCondition(30),
					unittest.GetCreatedCondition(60),
					unittest.GetCreatingCondition(90),
				},
				Versions: []infrastructurev1alpha2.CommonClusterStatusVersion{
					unittest.GetVersion(60, "8.7.6"),
					unittest.GetVersion(60, "8.7.5"),
				},
			},
			desiredVersion: "8.7.6",

			expectedResult: false,
		},
		// we can also upgrade multiple times
		{
			name: "case 5",

			status: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
					unittest.GetUpdatedCondition(10),
					unittest.GetUpdatingCondition(30),
					unittest.GetCreatedCondition(60),
					unittest.GetCreatingCondition(90),
				},
				Versions: []infrastructurev1alpha2.CommonClusterStatusVersion{
					unittest.GetVersion(10, "8.7.6"),
					unittest.GetVersion(60, "8.7.5"),
				},
			},
			desiredVersion: "8.7.7",

			expectedResult: true,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result := computeUpdatingCondition(tc.status, tc.desiredVersion)
			if result != tc.expectedResult {
				t.Fatalf("expected %v, got %v", tc.expectedResult, result)
			}
		})
	}
}

func TestComputeUpdatedCondition(t *testing.T) {
	testCases := []struct {
		name string

		status     infrastructurev1alpha2.CommonClusterStatus
		nodesReady bool

		expectedResult bool
	}{
		// the cluster is updating and nodes are ready
		{
			name: "case 0",

			status: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
					unittest.GetUpdatingCondition(30),
					unittest.GetCreatedCondition(60),
					unittest.GetCreatingCondition(90),
				},
			},
			nodesReady: true,

			expectedResult: true,
		},
		// the cluster is updating and nodes are not ready
		{
			name: "case 1",

			status: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
					unittest.GetUpdatingCondition(30),
					unittest.GetCreatedCondition(60),
					unittest.GetCreatingCondition(90),
				},
			},
			nodesReady: false,

			expectedResult: false,
		},
		// the cluster is already updated
		{
			name: "case 2",

			status: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
					unittest.GetUpdatedCondition(10),
					unittest.GetUpdatingCondition(30),
					unittest.GetCreatedCondition(60),
					unittest.GetCreatingCondition(90),
				},
			},
			nodesReady: true,

			expectedResult: false,
		},
		// the cluster is updating and nodes are ready for the second time
		{
			name: "case 3",

			status: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
					unittest.GetUpdatingCondition(10),
					unittest.GetUpdatedCondition(20),
					unittest.GetUpdatingCondition(40),
					unittest.GetCreatedCondition(60),
					unittest.GetCreatingCondition(90),
				},
			},
			nodesReady: true,

			expectedResult: true,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result := computeUpdatedCondition(tc.status, tc.nodesReady)
			if result != tc.expectedResult {
				t.Fatalf("expected %v, got %v", tc.expectedResult, result)
			}
		})
	}
}

func TestComputeVersionChange(t *testing.T) {
	testCases := []struct {
		name string

		status         infrastructurev1alpha2.CommonClusterStatus
		nodesReady     bool
		desiredVersion string

		expectedResult bool
	}{
		// the cluster has transitioned, nodes are ready but the version is not set yet
		{
			name: "case 0",

			status: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
					unittest.GetUpdatedCondition(10),
					unittest.GetUpdatingCondition(30),
					unittest.GetCreatedCondition(60),
					unittest.GetCreatingCondition(90),
				},
				Versions: []infrastructurev1alpha2.CommonClusterStatusVersion{
					unittest.GetVersion(60, "8.7.5"),
				},
			},
			nodesReady:     true,
			desiredVersion: "8.7.6",

			expectedResult: true,
		},
		// the cluster has not transitioned yet
		{
			name: "case 1",

			status: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
					unittest.GetCreatingCondition(90),
				},
				Versions: []infrastructurev1alpha2.CommonClusterStatusVersion{},
			},
			nodesReady:     true,
			desiredVersion: "8.7.5",

			expectedResult: false,
		},
		// the nodes are not ready yet
		{
			name: "case 2",

			status: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
					unittest.GetUpdatedCondition(10),
					unittest.GetUpdatingCondition(30),
					unittest.GetCreatedCondition(60),
					unittest.GetCreatingCondition(90),
				},
				Versions: []infrastructurev1alpha2.CommonClusterStatusVersion{
					unittest.GetVersion(60, "8.7.5"),
				},
			},
			nodesReady:     false,
			desiredVersion: "8.7.6",

			expectedResult: false,
		},
		// the version is already set
		{
			name: "case 3",

			status: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
					unittest.GetUpdatedCondition(10),
					unittest.GetUpdatingCondition(30),
					unittest.GetCreatedCondition(60),
					unittest.GetCreatingCondition(90),
				},
				Versions: []infrastructurev1alpha2.CommonClusterStatusVersion{
					unittest.GetVersion(10, "8.7.6"),
					unittest.GetVersion(60, "8.7.5"),
				},
			},
			nodesReady:     true,
			desiredVersion: "8.7.6",

			expectedResult: false,
		},
		// the cluster has not transitioned yet
		{
			name: "case 4",

			status: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
					unittest.GetUpdatingCondition(30),
					unittest.GetCreatedCondition(60),
					unittest.GetCreatingCondition(90),
				},
				Versions: []infrastructurev1alpha2.CommonClusterStatusVersion{
					unittest.GetVersion(60, "8.7.5"),
				},
			},
			nodesReady:     true,
			desiredVersion: "8.7.6",

			expectedResult: false,
		},
		// we roll back
		{
			name: "case 5",

			status: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
					unittest.GetUpdatedCondition(5),
					unittest.GetUpdatingCondition(15),
					unittest.GetUpdatedCondition(20),
					unittest.GetUpdatingCondition(40),
					unittest.GetCreatedCondition(60),
					unittest.GetCreatingCondition(90),
				},
				Versions: []infrastructurev1alpha2.CommonClusterStatusVersion{
					unittest.GetVersion(20, "8.7.6"),
					unittest.GetVersion(60, "8.7.5"),
				},
			},
			nodesReady:     true,
			desiredVersion: "8.7.5",

			expectedResult: true,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result := computeVersionChange(tc.status, tc.nodesReady, tc.desiredVersion)
			if result != tc.expectedResult {
				t.Fatalf("expected %v, got %v", tc.expectedResult, result)
			}
		})
	}
}
