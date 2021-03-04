package key

// AppSpec is used to define app custom resources.
type AppSpec struct {
	App     string
	AppName string
	Catalog string
	Chart   string
	// Whether app is installed for clusterapi clusters only.
	ClusterAPIOnly bool
	// ConfigMapName overrides the name, otherwise the cluster values configmap
	// is used.
	ConfigMapName string
	// Whether app is installed for legacy clusters only.
	// InCluster determines if the app CR should use in cluster. Otherwise the
	// cluster kubeconfig is specified.
	InCluster       bool
	LegacyOnly      bool
	Namespace       string
	UseUpgradeForce bool
	Version         string
}

// ChartSpec is used to define chartconfig custom resources.
type ChartSpec struct {
	AppName           string
	ChartName         string
	ConfigMapName     string
	UserConfigMapName string
}
