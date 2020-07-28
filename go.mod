module github.com/giantswarm/cluster-operator

go 1.14

require (
	github.com/Masterminds/semver v1.5.0
	github.com/ghodss/yaml v1.0.0
	github.com/giantswarm/apiextensions v0.4.17-0.20200723160042-89aed92d1080
	github.com/giantswarm/certs/v2 v2.0.1-0.20200714195905-72e095f60587
	github.com/giantswarm/errors v0.2.3
	github.com/giantswarm/exporterkit v0.2.0
	github.com/giantswarm/k8sclient/v3 v3.1.3-0.20200724085258-345602646ea8
	github.com/giantswarm/kubeconfig v0.2.2-0.20200724082502-5a2c86aaf684
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.2.1
	github.com/giantswarm/microkit v0.2.1
	github.com/giantswarm/micrologger v0.3.1
	github.com/giantswarm/operatorkit v1.2.1-0.20200724133006-e6de285a86c0
	github.com/giantswarm/resource v0.2.1-0.20200724133802-35859e18b11d
	github.com/giantswarm/tenantcluster/v2 v2.0.1-0.20200724133643-1c49720f2600
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.7.1
	github.com/spf13/afero v1.3.2
	github.com/spf13/viper v1.7.0
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.5
	k8s.io/apiextensions-apiserver v0.18.5
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v0.18.5
	sigs.k8s.io/cluster-api v0.3.7
	sigs.k8s.io/controller-runtime v0.6.1
)

replace sigs.k8s.io/cluster-api v0.3.7 => github.com/giantswarm/cluster-api v0.3.8-0.20200723145930-f76c9cd8e8d1
