package project

var (
	bundleVersion = "2.0.2-dev-jhbok"
	description   = "The cluster-operator manages Kubernetes tenant cluster resources."
	gitSHA        = "n/a"
	name          = "cluster-operator"
	source        = "https://github.com/giantswarm/cluster-operator"
	version       = "n/a"
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
