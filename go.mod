module github.com/giantswarm/cluster-operator

go 1.16

require (
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/giantswarm/apiextensions v0.2.0
	github.com/giantswarm/apprclient v0.0.0-20191209123802-955b7e89e6e2
	github.com/giantswarm/backoff v0.2.0
	github.com/giantswarm/certs/v2 v2.0.0
	github.com/giantswarm/clusterclient v0.0.0-20200127145418-6c6f565f94c7
	github.com/giantswarm/errors v0.0.0-20200227191412-38679efaafe8
	github.com/giantswarm/k8sclient v0.2.0
	github.com/giantswarm/kubeconfig v0.0.0-20191209121754-c5784ae65a49
	github.com/giantswarm/microclient v0.0.0-20190809131213-459b479d046f // indirect
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.2.0
	github.com/giantswarm/microkit v0.2.0
	github.com/giantswarm/micrologger v0.3.1
	github.com/giantswarm/operatorkit v0.2.1
	github.com/giantswarm/resource v0.0.0-20201203121554-9ca9ddbde7cf
	github.com/giantswarm/tenantcluster/v2 v2.0.0
	github.com/giantswarm/versionbundle v0.2.0
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.3.3 // indirect
	github.com/google/go-cmp v0.4.0
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/common v0.9.1 // indirect
	github.com/spf13/afero v1.2.2
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.6.2
	golang.org/x/crypto v0.0.0-20200210222208-86ce3cb69678 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20200202164722-d101bd2416d5 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	gopkg.in/ini.v1 v1.52.0 // indirect
	gopkg.in/resty.v1 v1.12.0
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.16.6
	k8s.io/apimachinery v0.16.6
	k8s.io/client-go v0.16.6
	k8s.io/kube-openapi v0.0.0-20200204173128-addea2498afe // indirect
	k8s.io/utils v0.0.0-20200124190032-861946025e34 // indirect
)

// replace gopkg.in/fsnotify.v1 v1.4.7 => gopkg.in/fsnotify/fsnotify.v1 v1.4.7
