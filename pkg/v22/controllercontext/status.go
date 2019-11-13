package controllercontext

type ContextStatus struct {
	Worker ContextStatusWorker
}

type ContextStatusWorker struct {
	Nodes int32
}
