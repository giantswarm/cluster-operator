package key

const (
	LabelOperatorVersion = "aws-operator.giantswarm.io/version"
)

func OperatorVersion(getter LabelsGetter) string {
	return getter.GetLabels()[LabelOperatorVersion]
}
