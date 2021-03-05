package label

const (
	// App label is deprecated and is replaced by AppKubernetesName.
	App = "app"

	// AppKubernetesName is a standard label for Kubernetes resources.
	AppKubernetesName = "app.kubernetes.io/name"

	// AppOperatorHelmMajorVersion is a label for chart-operator app CRs.
	AppOperatorHelmMajorVersion = "app-operator.giantswarm.io/helm-major-version"
)
