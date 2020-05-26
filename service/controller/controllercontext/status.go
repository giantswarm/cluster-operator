package controllercontext

type ContextStatus struct {
	// Worker is a map of key value pairs where the key is the machine deployment
	// ID. The map value is a structure holding node information for the
	// corresponding machine deployment.
	Worker map[string]ContextStatusWorker
}

type ContextStatusWorker struct {
	Nodes int32
	Ready int32
}
