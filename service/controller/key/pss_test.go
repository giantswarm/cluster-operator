package key

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/cluster-operator/v5/pkg/label"
)

func TestIsPSSRelease(t *testing.T) {
	tests := []struct {
		name    string
		getter  LabelsGetter
		want    bool
		wantErr bool
	}{
		{
			name:    "= 19.3.0",
			getter:  getClusterWithLabelVersion("19.3.0"),
			want:    true,
			wantErr: false,
		},
		{
			name:    "< 19.3.0",
			getter:  getClusterWithLabelVersion("19.2.999999"),
			want:    false,
			wantErr: false,
		},
		{
			name:    "19.3.0 test version",
			getter:  getClusterWithLabelVersion("19.3.0-blah"),
			want:    true,
			wantErr: false,
		},
		{
			name:    "> 19.3.0",
			getter:  getClusterWithLabelVersion("19.3.1"),
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsPSSRelease(tt.getter)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsPSSRelease() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsPSSRelease() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func getClusterWithLabelVersion(version string) *v1beta1.Cluster {
	return &v1beta1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				label.ReleaseVersion: version,
			},
		},
	}
}
