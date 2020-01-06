package project

var (
	bundleVersion = "2.0.1-xh3b4sd"
	description   = "The cluster-operator manages Kubernetes guest cluster resources."
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
