module github.com/giantswarm/cluster-operator

go 1.14

require (
	github.com/Masterminds/semver v1.5.0
	github.com/ghodss/yaml v1.0.0
	github.com/giantswarm/apiextensions v0.4.4
	github.com/giantswarm/certs/v2 v2.0.0
	github.com/giantswarm/errors v0.2.2
	github.com/giantswarm/exporterkit v0.2.0
	github.com/giantswarm/k8sclient/v3 v3.1.0
	github.com/giantswarm/kubeconfig v0.2.0
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.2.0
	github.com/giantswarm/microkit v0.2.1
	github.com/giantswarm/micrologger v0.3.1
	github.com/giantswarm/operatorkit v1.0.0
	github.com/giantswarm/resource v0.2.0
	github.com/giantswarm/tenantcluster/v2 v2.0.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.5.1
	github.com/spf13/afero v1.2.2
	github.com/spf13/viper v1.6.3
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.17.2
	k8s.io/apiextensions-apiserver v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	sigs.k8s.io/cluster-api v0.3.6
	sigs.k8s.io/controller-runtime v0.5.2
)
