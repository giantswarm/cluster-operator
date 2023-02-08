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
	// DependsOn list of dependencies of this app.
	DependsOn []string
	// InCluster determines if the app CR should use in cluster. Otherwise the
	// cluster kubeconfig is specified.
	InCluster bool
	// Whether app is installed for legacy clusters only.
	LegacyOnly      bool
	Namespace       string
	UseUpgradeForce bool
	Version         string
}

func (a AppSpec) GetAppName() string {
	if a.AppName != "" {
		return a.AppName
	}
	return a.App
}
