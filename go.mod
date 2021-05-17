module github.com/giantswarm/cluster-operator/v3

go 1.14

require (
	github.com/ghodss/yaml v1.0.0
	github.com/giantswarm/apiextensions/v3 v3.23.0
	github.com/giantswarm/backoff v0.2.0
	github.com/giantswarm/certs/v3 v3.1.1
	github.com/giantswarm/errors v0.3.0
	github.com/giantswarm/exporterkit v0.2.1
	github.com/giantswarm/k8sclient/v5 v5.11.0
	github.com/giantswarm/kubeconfig/v2 v2.0.0
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/microkit v0.2.2
	github.com/giantswarm/micrologger v0.5.0
	github.com/giantswarm/operatorkit/v4 v4.3.1
	github.com/giantswarm/resource/v2 v2.3.0
	github.com/giantswarm/tenantcluster/v3 v3.0.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.10.0
	github.com/spf13/afero v1.6.0
	github.com/spf13/viper v1.7.1
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.18.18
	k8s.io/apiextensions-apiserver v0.18.9
	k8s.io/apimachinery v0.18.18
	k8s.io/client-go v0.18.18
	sigs.k8s.io/cluster-api v0.3.16
	sigs.k8s.io/controller-runtime v0.6.4
)

replace (
	github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	sigs.k8s.io/cluster-api => github.com/giantswarm/cluster-api v0.3.13-gs
)
