package controllercontext

type ContextStatus struct {
	Endpoint ContextStatusEndpoint
	// Versions is a map of key value pairs where the map key is a version label
	// of a given operator. The map value is the version of the corresponding
	// operator. See also the operatorversions resource.
	//
	//     aws-operator.giantswarm.io/version: 6.5.0
	//
	Versions map[string]string
	// Worker is a map of key value pairs where the key is the machine deployment
	// ID. The map value is a structure holding node information for the
	// corresponding machine deployment.
	Worker map[string]ContextStatusWorker
}

type ContextStatusEndpoint struct {
	Base string
}

type ContextStatusWorker struct {
	Nodes int32
	Ready int32
}
