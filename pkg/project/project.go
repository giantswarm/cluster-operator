package project

var (
	description = "The cluster-operator manages Kubernetes tenant cluster resources."
	gitSHA      = "n/a"
	name        = "cluster-operator"
	source      = "https://github.com/giantswarm/cluster-operator"
	version     = "999.9.9"
)

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
