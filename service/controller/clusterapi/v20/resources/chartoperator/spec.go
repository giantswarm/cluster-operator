package chartoperator

// Image holds the image settings for chart-operator chart.
type Image struct {
	Registry string `json:"registry"`
}

// ResourceState holds the state of the chart to be reconciled.
type ResourceState struct {
	// ChartName is the name of the Helm Chart.
	// e.g. chart-operator-chart
	ChartName string
	// ChartValues are the values to be passed to the chart-operator Helm Chart.
	// e.g. chart-operator-chart
	ChartValues Values
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

// Tiller holds the Tiller settings for chart-operator chart.
type Tiller struct {
	Namespace string `json:"namespace"`
}

// Values represents the values to be passed to Helm commands related to
// chart-operator chart.
type Values struct {
	ClusterDNSIP string `json:"clusterDNSIP"`
	Image        Image  `json:"image"`
	Tiller       Tiller `json:"tiller"`
}
