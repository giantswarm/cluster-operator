module github.com/giantswarm/cluster-operator/v3

go 1.14

require (
	github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/giantswarm/apiextensions/v3 v3.13.0
	github.com/giantswarm/certs/v3 v3.1.0
	github.com/giantswarm/errors v0.2.3
	github.com/giantswarm/exporterkit v0.2.0
	github.com/giantswarm/k8sclient/v5 v5.0.0
	github.com/giantswarm/kubeconfig/v2 v2.0.0
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/microkit v0.2.2
	github.com/giantswarm/micrologger v0.4.0
	github.com/giantswarm/operatorkit/v4 v4.0.0
	github.com/giantswarm/resource/v2 v2.2.1-0.20201209145524-83246f56d150
	github.com/giantswarm/tenantcluster/v3 v3.0.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.8.0
	github.com/spf13/afero v1.5.0
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1 // indirect
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.18.9
	k8s.io/apiextensions-apiserver v0.18.9
	k8s.io/apimachinery v0.18.9
	k8s.io/client-go v0.18.9
	sigs.k8s.io/cluster-api v0.3.11
	sigs.k8s.io/controller-runtime v0.6.4
)

// keep in sync with giantswarm/apiextensions
replace sigs.k8s.io/cluster-api => github.com/giantswarm/cluster-api v0.3.10-gs
