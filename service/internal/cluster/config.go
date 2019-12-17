package cluster

import "net"

// Config data structure provides configuration values for cluster.
type Config struct {
	ClusterID    string
	CertTTL      string
	Domain       Domain
	IP           IP
	Organization string
}

// Domain data structure holds different domain entries for cluster components.
type Domain struct {
	API                string
	Calico             string
	CalicoEtcdClient   string
	Etcd               string
	FlanneldEtcdClient string
	InternalAPI        string
	NodeOperator       string
	Prometheus         string
	ServiceAccount     string
	Worker             string
}

// IP data structure holds IP entries for cluster components.
type IP struct {
	API   net.IP
	Range net.IP
}