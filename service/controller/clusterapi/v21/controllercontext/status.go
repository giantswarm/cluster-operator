package controllercontext

type ContextStatus struct {
	// Versions is a map of key value pairs where the map key is a version label
	// of a given operator. The map value is the version of the corresponding
	// operator.
	Versions map[string]string
	// Worker is a map of key value pairs where the key is the machine deployment
	// ID. The map value is a structure holding node information for the
	// corresponding machine deployment.
	Worker map[string]ContextStatusWorker
}

type ContextStatusWorker struct {
	Nodes int
	Ready int
}
