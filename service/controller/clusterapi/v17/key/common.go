package key

import "github.com/giantswarm/cluster-operator/pkg/label"

func OperatorVersion(getter LabelsGetter) string {
	return getter.GetLabels()[label.OperatorVersion]
}
