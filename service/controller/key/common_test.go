package key

import (
	"testing"

	"github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/v5/pkg/label"
)

// A mock object that implements LabelsGetter interface
type testObject struct {
	labels map[string]string
}

func (to *testObject) GetLabels() map[string]string {
	return to.labels
}

func Test_ClusterConfigMapName(t *testing.T) {
	testCases := []struct {
		description    string
		customObject   LabelsGetter
		expectedResult string
	}{
		{
			description:    "case 0: getting cluster configmap name",
			customObject:   &testObject{map[string]string{label.Cluster: "w7utg"}},
			expectedResult: "w7utg-cluster-values",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := ClusterConfigMapName(tc.customObject)
			if result != tc.expectedResult {
				t.Fatalf("expected ClusterConfigMapName %#q, got %#q", tc.expectedResult, result)
			}
		})
	}
}

func Test_ClusterID(t *testing.T) {
	testCases := []struct {
		description  string
		customObject LabelsGetter
		expectedID   string
	}{
		{
			description:  "empty value object produces empty ID",
			customObject: &testObject{},
			expectedID:   "",
		},
		{
			description:  "present ID value returned as ClusterID",
			customObject: &testObject{map[string]string{label.Cluster: "cluster-1"}},
			expectedID:   "cluster-1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			if cid := ClusterID(tc.customObject); cid != tc.expectedID {
				t.Fatalf("ClusterID %s doesn't match. expected: %s", cid, tc.expectedID)
			}
		})
	}
}

func TestIRSAEnabled(t *testing.T) {
	tests := []struct {
		name       string
		awsCluster *v1alpha3.AWSCluster
		want       bool
	}{
		{
			name:       "nil parameter",
			awsCluster: nil,
			want:       false,
		},
		{
			name:       "cluster with no release label",
			awsCluster: &v1alpha3.AWSCluster{},
			want:       false,
		},
		{
			name: "v18 cluster with no irsa annotation",
			awsCluster: &v1alpha3.AWSCluster{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"release.giantswarm.io/version": "18.0.0",
					},
					Annotations: nil,
				},
			},
			want: false,
		},
		{
			name: "v18 cluster with irsa annotation",
			awsCluster: &v1alpha3.AWSCluster{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"release.giantswarm.io/version": "18.0.0",
					},
					Annotations: map[string]string{
						"alpha.aws.giantswarm.io/iam-roles-for-service-accounts": "",
					},
				},
			},
			want: true,
		},
		{
			name: "v19 cluster with no irsa annotation",
			awsCluster: &v1alpha3.AWSCluster{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"release.giantswarm.io/version": "19.0.0",
					},
					Annotations: nil,
				},
			},
			want: true,
		},
		{
			name: "v19 cluster with no irsa annotation and wrong release version with leading 'v'",
			awsCluster: &v1alpha3.AWSCluster{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"release.giantswarm.io/version": "v19.0.0",
					},
					Annotations: nil,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IRSAEnabled(tt.awsCluster); got != tt.want {
				t.Errorf("IRSAEnabled(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
