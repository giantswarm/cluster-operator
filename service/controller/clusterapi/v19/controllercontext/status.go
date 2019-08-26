package controllercontext

type ContextStatus struct {
	Versions map[string]string
	Worker   ContextStatusWorker
}

type ContextStatusWorker struct {
	Nodes int
	Ready int
}
