package chart

// State holds the state of the chart to be reconciled.
type State struct {
	// ChartName is the name of the Helm Chart.
	// e.g. chart-operator-chart
	ChartName string
	// ReleaseName is the name of the Helm release when the chart is deployed.
	// e.g. chart-operator
	ReleaseName string
	// ReleaseStatus is the status of the Helm Release.
	// e.g. DEPLOYED
	ReleaseStatus string
	// ReleaseVersion is the version of the Helm Chart to be deployed.
	// e.g. 0.1.2
	ReleaseVersion string
}
