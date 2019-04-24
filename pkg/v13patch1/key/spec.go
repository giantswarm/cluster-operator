package key

// ChartSpec is used to define chartconfig custom resources.
type ChartSpec struct {
	AppName           string
	ChannelName       string
	ChartName         string
	ConfigMapName     string
	Namespace         string
	ReleaseName       string
	UserConfigMapName string
}
