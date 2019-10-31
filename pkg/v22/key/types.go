package key

// AppSpec is used to define app custom resources.
type AppSpec struct {
	App             string
	Catalog         string
	Chart           string
	ClusterAPI      bool
	Namespace       string
	UseUpgradeForce bool
	Version         string
}

// ChartSpec is used to define chartconfig custom resources.
type ChartSpec struct {
	AppName           string
	ChannelName       string
	ChartName         string
	ConfigMapName     string
	HasAppCR          bool
	Namespace         string
	ReleaseName       string
	UseUpgradeForce   bool
	UserConfigMapName string
}
