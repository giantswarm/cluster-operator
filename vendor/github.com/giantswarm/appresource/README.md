[![CircleCI](https://circleci.com/gh/giantswarm/appresource.svg?style=shield)](https://circleci.com/gh/giantswarm/appresource)

# appresource

Package appresource implements a generic [operatorkit] resource for managing
[app] custom resources. The custom resources are managed by [app-operator]
which installs applications packaged as Helm charts into Kubernetes clusters.

## Example custom resource

The app custom resources look something like this.

```yaml
apiVersion: application.giantswarm.io/v1alpha1
kind: App
metadata:
name: "prometheus"
labels:
    app-operator.giantswarm.io/version: "1.0.0"

spec:
  catalog: "giantswarm"
  name: "kube-state-metrics"
  namespace: "kube-system"
  version: "0.4.0"

  config:
    configMap:
      name: "eggs2-cluster-values"
      namespace: "eggs2"
  kubeConfig:
    context:
      name: "eggs2"
    secret:
      name: "eggs2-kubeconfig"
      namespace: "eggs2"

  status:
    appVersion: "1.7.2"
    release:
      lastDeployed: "2019-11-30T21:06:20Z"
      status: "DEPLOYED"
    version: "0.4.0"
```

## Used by

- [cluster-operator]
- [release-operator]

## License

appresource is under the Apache 2.0 license. See the [LICENSE](LICENSE) file
for details.

[app]: https: github.com/giantswarm/apiextensions/blob/master/pkg/apis/application/v1alpha1/app_types.go
[app-operator]: https: github.com/giantswarm/app-operator
[cluster-operator]: https: github.com/giantswarm/app-operator
[operatorkit]: https: github.com/giantswarm/operatorkit
[release-operator]: https: github.com/giantswarm/release-operator
