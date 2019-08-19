package key

import (
	"testing"
	"time"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func Test_AWSClusterStatusAccessor(t *testing.T) {
	var accessor AWSClusterStatusAccessor
	var cluster cmav1alpha1.Cluster

	preSetClusterID := "abc12"
	preSetClusterVersion := "1.2.3"

	{
		status := accessor.GetCommonClusterStatus(cluster)
		status.ID = preSetClusterID
		cluster = accessor.SetCommonClusterStatus(cluster, status)
	}

	{
		status := accessor.GetCommonClusterStatus(cluster)
		if status.ID != preSetClusterID {
			t.Fatalf("expected cluster ID %s, got %s", preSetClusterID, status.ID)
		}
	}

	{
		status := accessor.GetCommonClusterStatus(cluster)
		if len(status.Versions) != 0 {
			t.Fatalf("expected cluster.Versions to be empty, found %d versions: %#v", len(status.Versions), status.Versions)
		}

		newVer := g8sv1alpha1.CommonClusterStatusVersion{
			LastTransitionTime: g8sv1alpha1.DeepCopyTime{Time: time.Now()},
			Version:            preSetClusterVersion,
		}
		status.Versions = append(status.Versions, newVer)

		cluster = accessor.SetCommonClusterStatus(cluster, status)
	}

	{
		status := accessor.GetCommonClusterStatus(cluster)
		if len(status.Versions) != 1 {
			t.Fatalf("expected cluster.Versions to have exactly one version, found %d versions: %#v", len(status.Versions), status.Versions)
		}

		if status.Versions[0].Version != preSetClusterVersion {
			t.Fatalf("expected cluster version to be %s, got %s", preSetClusterVersion, status.Versions[0].Version)
		}
	}
}
