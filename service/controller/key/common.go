package key

import (
	"fmt"
	"strings"

	"github.com/blang/semver"

	"github.com/giantswarm/cluster-operator/v5/pkg/label"
)

const (
	IRSAAppName     = "aws-pod-identity-webhook"
	IRSAAppCatalog  = "default"
	IRSAAppVersion  = "0.3.1"
	V19AlphaRelease = "19.0.0-alpha1"
)

func APISecretName(getter LabelsGetter) string {
	return fmt.Sprintf("%s-api", ClusterID(getter))
}

// ClusterConfigMapName returns the cluster name used in the configMap
// generated for this tenant cluster.
func ClusterConfigMapName(getter LabelsGetter) string {
	return fmt.Sprintf("%s-cluster-values", ClusterID(getter))
}

func ClusterID(getter LabelsGetter) string {
	return getter.GetLabels()[label.Cluster]
}

func IsDeleted(getter DeletionTimestampGetter) bool {
	return getter.GetDeletionTimestamp() != nil
}

func IsV19Release(releaseVersion *semver.Version) bool {
	v19, _ := semver.New(V19AlphaRelease)
	return releaseVersion.Major >= v19.Major
}

func KubeConfigClusterName(getter LabelsGetter) string {
	return fmt.Sprintf("giantswarm-%s", ClusterID(getter))
}

func KubeConfigSecretName(getter LabelsGetter) string {
	return fmt.Sprintf("%s-kubeconfig", ClusterID(getter))
}

func MachineDeployment(getter LabelsGetter) string {
	return getter.GetLabels()[label.MachineDeployment]
}

func OperatorVersion(getter LabelsGetter) string {
	return getter.GetLabels()[label.OperatorVersion]
}

func OrganizationID(getter LabelsGetter) string {
	return getter.GetLabels()[label.Organization]
}

func ReleaseName(releaseVersion string) string {
	return fmt.Sprintf("v%s", releaseVersion)
}

func ReleaseVersion(getter LabelsGetter) string {
	return getter.GetLabels()[label.ReleaseVersion]
}

func IsBundle(appName string) bool {
	return strings.HasSuffix(appName, "-bundle")
}
