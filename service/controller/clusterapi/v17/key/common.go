package key

import "github.com/giantswarm/cluster-operator/pkg/label"

func ClusterID(getter LabelsGetter) string {
	return getter.GetLabels()[label.Cluster]
}

func MachineDeployment(getter LabelsGetter) string {
	return getter.GetLabels()[label.MachineDeployment]
}

func OperatorVersion(getter LabelsGetter) string {
	return getter.GetLabels()[label.OperatorVersion]
}

func ReleaseVersion(getter LabelsGetter) string {
	return getter.GetLabels()[label.ReleaseVersion]
}
