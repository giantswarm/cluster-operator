package controllercontext

type ContextStatus struct {
	TenantCluster ContextStatusTenantCluster
	Worker        ContextStatusWorker
}

type ContextStatusTenantCluster struct {
	IsUnavailable bool
}

type ContextStatusWorker struct {
	Nodes int
}
