package collector

import (
	"strconv"
	"testing"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"

	"github.com/giantswarm/cluster-operator/v3/service/internal/unittest"
)

func TestCollectClusterTransition(t *testing.T) {
	testCases := []struct {
		name   string
		status infrastructurev1alpha3.CommonClusterStatus

		expectCreated int
		expectUpdated int
	}{
		// the cluster is creating
		{
			name: "case 0",
			status: infrastructurev1alpha3.CommonClusterStatus{
				Conditions: []infrastructurev1alpha3.CommonClusterStatusCondition{
					unittest.GetCreatingCondition(20),
				},
			},

			expectCreated: 0,
			expectUpdated: 0,
		},
		// the cluster is created
		{
			name: "case 1",
			status: infrastructurev1alpha3.CommonClusterStatus{
				Conditions: []infrastructurev1alpha3.CommonClusterStatusCondition{
					unittest.GetCreatedCondition(0),
					unittest.GetCreatingCondition(30),
				},
			},

			expectCreated: 30*60 - 1,
			expectUpdated: 0,
		},
		// the cluster is updating
		{
			name: "case 2",
			status: infrastructurev1alpha3.CommonClusterStatus{
				Conditions: []infrastructurev1alpha3.CommonClusterStatusCondition{
					unittest.GetUpdatingCondition(30),
					unittest.GetCreatedCondition(60),
					unittest.GetCreatingCondition(90),
				},
			},

			expectCreated: 30*60 - 1,
			expectUpdated: 0,
		},
		// the cluster is updated
		{
			name: "case 3",
			status: infrastructurev1alpha3.CommonClusterStatus{
				Conditions: []infrastructurev1alpha3.CommonClusterStatusCondition{
					unittest.GetUpdatedCondition(10),
					unittest.GetUpdatingCondition(30),
					unittest.GetCreatedCondition(60),
					unittest.GetCreatingCondition(90),
				},
			},

			expectCreated: 30*60 - 1,
			expectUpdated: 20*60 - 1,
		},
		// the cluster is updating for the second time
		{
			name: "case 4",
			status: infrastructurev1alpha3.CommonClusterStatus{
				Conditions: []infrastructurev1alpha3.CommonClusterStatusCondition{
					unittest.GetUpdatingCondition(10),
					unittest.GetUpdatedCondition(20),
					unittest.GetUpdatingCondition(40),
					unittest.GetCreatedCondition(60),
					unittest.GetCreatingCondition(90),
				},
			},
			expectCreated: 30*60 - 1,
			expectUpdated: 0,
		},
		// the cluster is updated for the second time
		{
			name: "case 5",
			status: infrastructurev1alpha3.CommonClusterStatus{
				Conditions: []infrastructurev1alpha3.CommonClusterStatusCondition{
					unittest.GetUpdatedCondition(5),
					unittest.GetUpdatingCondition(10),
					unittest.GetUpdatedCondition(20),
					unittest.GetUpdatingCondition(40),
					unittest.GetCreatedCondition(60),
					unittest.GetCreatingCondition(90),
				},
			},

			expectCreated: 30*60 - 1,
			expectUpdated: 5*60 - 1,
		},
		// stuck creating
		{
			name: "case 6",
			status: infrastructurev1alpha3.CommonClusterStatus{
				Conditions: []infrastructurev1alpha3.CommonClusterStatusCondition{
					unittest.GetCreatingCondition(90),
				},
			},

			expectCreated: 999999999999,
			expectUpdated: 0,
		},
		// stuck updating
		{
			name: "case 7",
			status: infrastructurev1alpha3.CommonClusterStatus{
				Conditions: []infrastructurev1alpha3.CommonClusterStatusCondition{
					unittest.GetUpdatingCondition(200),
					unittest.GetCreatedCondition(300),
					unittest.GetCreatingCondition(330),
				},
			},

			expectCreated: 30*60 - 1,
			expectUpdated: 999999999999,
		},
		// the cluster is stuck updating for the second timne
		{
			name: "case 8",
			status: infrastructurev1alpha3.CommonClusterStatus{
				Conditions: []infrastructurev1alpha3.CommonClusterStatusCondition{
					unittest.GetUpdatingCondition(180),
					unittest.GetUpdatedCondition(200),
					unittest.GetUpdatingCondition(220),
					unittest.GetCreatedCondition(300),
					unittest.GetCreatingCondition(330),
				},
			},

			expectCreated: 30*60 - 1,
			expectUpdated: 999999999999,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, created := getCreateMetrics(tc.status)
			if int(created) != tc.expectCreated {
				t.Fatalf("expected %v, got %v", tc.expectCreated, int(created))
			}
			_, updated := getUpdateMetrics(tc.status)
			if int(updated) != tc.expectUpdated {
				t.Fatalf("expected %v, got %v", tc.expectUpdated, int(updated))
			}
		})
	}
}
