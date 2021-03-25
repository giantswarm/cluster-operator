package project

var (
	bundleVersion = "0.26.0"
	description   = "The cluster-operator manages Kubernetes tenant cluster resources."
	gitSHA        = "n/a"
	name          = "cluster-operator"
	source        = "https://github.com/giantswarm/cluster-operator"
	version       = "0.26.0"
)

func BundleVersion() string {
	return bundleVersion
}

func Description() string {
	return description
}

func GitSHA() string {
	return gitSHA
}

func Name() string {
	return name
}

func Source() string {
	return source
}

func Version() string {
	return version
}
