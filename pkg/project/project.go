package project

var (
	description = "The cluster-operator manages Kubernetes tenant cluster resources."
	gitSHA      = "n/a"
	name        = "cluster-operator"
	source      = "https://github.com/giantswarm/cluster-operator"
	// version     = "5.8.1-dev"
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
	return "5.8.0"
}
