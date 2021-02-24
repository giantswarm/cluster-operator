module github.com/giantswarm/cluster-operator

go 1.16

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/giantswarm/apiextensions v0.1.1
	github.com/giantswarm/apprclient v0.0.0-20191209123802-955b7e89e6e2
	github.com/giantswarm/backoff v0.0.0-20200209120535-b7cb1852522d
	github.com/giantswarm/certs v0.0.0-20191209164338-3f774da07e59
	github.com/giantswarm/clusterclient v0.0.0-20200127145418-6c6f565f94c7
	github.com/giantswarm/errors v0.0.0-20200227191412-38679efaafe8
	github.com/giantswarm/exporterkit v0.0.0-20190619131829-9749deade60f // indirect
	github.com/giantswarm/k8sclient v0.0.0-20200120104955-1542917096d6
	github.com/giantswarm/kubeconfig v0.0.0-20191209121754-c5784ae65a49
	github.com/giantswarm/microclient v0.0.0-20190809131213-459b479d046f // indirect
	github.com/giantswarm/microendpoint v0.0.0-20191121160659-e991deac2653
	github.com/giantswarm/microerror v0.1.0
	github.com/giantswarm/microkit v0.0.0-20191023091504-429e22e73d3e
	github.com/giantswarm/micrologger v0.1.1
	github.com/giantswarm/operatorkit v0.0.0-20200309100035-748c399a1bba
	github.com/giantswarm/resource v0.0.0-20201203121554-9ca9ddbde7cf
	github.com/giantswarm/tenantcluster v0.0.0-20191209135534-800974a4d4bf
	github.com/giantswarm/to v0.2.0 // indirect
	github.com/giantswarm/versionbundle v0.0.0-20200203095303-cd94540b7d5a
	github.com/go-logfmt/logfmt v0.5.0 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.3.3 // indirect
	github.com/google/go-cmp v0.3.1
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/juju/errgo v0.0.0-20140925100237-08cceb5d0b53 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/common v0.9.1 // indirect
	github.com/prometheus/procfs v0.0.8 // indirect
	github.com/spf13/afero v1.2.2
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.6.2
	golang.org/x/crypto v0.0.0-20200210222208-86ce3cb69678 // indirect
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20200202164722-d101bd2416d5 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	gopkg.in/ini.v1 v1.52.0 // indirect
	gopkg.in/resty.v1 v1.12.0
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.0.0-20191114100352-16d7abae0d2a
	k8s.io/apiextensions-apiserver v0.0.0-20191114105449-027877536833 // indirect
	k8s.io/apimachinery v0.16.5-beta.1
	k8s.io/client-go v0.0.0-20191114101535-6c5935290e33
	k8s.io/klog v1.0.0 // indirect
	k8s.io/kube-openapi v0.0.0-20200204173128-addea2498afe // indirect
	k8s.io/utils v0.0.0-20200124190032-861946025e34 // indirect
	sigs.k8s.io/controller-runtime v0.4.0 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

// replace gopkg.in/fsnotify.v1 v1.4.7 => gopkg.in/fsnotify/fsnotify.v1 v1.4.7
