package controllercontext

type ContextStatus struct {
	// Apps is a slice of the apps and versions that should be created for a specific release.
	// It is fetched from cluster-service by the releaseversions resource.
	//
	//     - coredns: 1.15.0
	//
	Apps []App
	// Versions is a map of key value pairs where the map key is a version label
	// of a given operator. The map value is the version of the corresponding
	// operator. See also the releaseversions resource.
	//
	//     aws-operator.giantswarm.io/version: 6.5.0
	//
	Versions map[string]string
	// Worker is a map of key value pairs where the key is the machine deployment
	// ID. The map value is a structure holding node information for the
	// corresponding machine deployment.
	Worker map[string]ContextStatusWorker
	// Master is a map of key value pairs where the key is the control plane
	// ID. The map value is a structure holding node information for the
	// corresponding control plane.
	Master map[string]ContextStatusMaster
}

type App struct {
	App              string
	ComponentVersion string
	Version          string
}

type ContextStatusWorker struct {
	Nodes int32
	Ready int32
}

type ContextStatusMaster struct {
	Nodes int32
	Ready int32
}
