# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- Fix user config CM mapping for bundle apps.

### Added

- Read app dependencies from Release CR to avoid deadlock installing apps in new clusters.

## [5.4.0] - 2023-01-30

### Added

- Add `aws.region` field in the cluster configmap.

### Changed

- Make the `aws` related fields only present on aws clusters' cluster configmap.

## [5.3.0] - 2022-11-03

### Changed

- Enable IRSA by default on v19+ clusters.

## [5.2.0] - 2022-10-10

### Added

- Allow disabling cilium's kube-proxy replacement feature by adding an annotation to the Cluster CR. 

## [5.1.0] - 2022-10-01

### Added

- Support for App bundles in default apps.

### Changed

- Bump `apiextensions-application` to the `0.6.0` version.

## [5.0.0] - 2022-09-26

### Changed

- Enable kube-proxy replacement mode in Cilium app.
- Enable bootstrap mode for chart operator.

### Added

- Add `vpc ID` to WC configmap on AWS.
- Add support for extraConfigs field in App CR.

## [4.6.2] - 2022-09-12

### Fixed

- Use `AzureConfig`'s `Spec.Azure.VirtualNetwork.CalicoSubnetCIDR` field for Calico CIDR rather than `Spec.Cluster.Calico.Subnet`.

## [4.6.1] - 2022-08-31

### Changed

- Empty release to fix broken automation.

## [4.6.0] - 2022-08-31

### Fixed

- Fixed finding of apps with and without the -app suffix in catalogs.

## [4.5.2] - 2022-08-11

### Changed

- Set `cni.exclusive` to `false` in cilium app config.

## [4.5.1] - 2022-08-09

### Changed

- Add `CNI_CONF_NAME` env to cilium app config.

## [4.5.0] - 2022-08-08

### Added

- Add cilium app config map.

## [4.4.0] - 2022-07-21

### Changed

- Set `chartOperator.cni.install` to true to allow installing CNI as app.

## [4.3.0] - 2022-06-02

### Changed

- Do not update "app-operator.giantswarm.io/version" label on app-operators when their value is 0.0.0 (aka they are reconciled by the management cluster app-operator). This is a use-case for App Bundles for example, because the App CRs they contain should be created in the MC so should be reconciled by the MC app-operator.

## [4.2.0] - 2022-05-25

### Added

- Add cluster values for IRSA.

## [4.1.0] - 2022-04-27

### Changed

- Store kubeconfig copy in `.data.value` field of the Secret.

## [4.0.2] - 2022-04-14

### Fixed

- List apps by namespace.

## [4.0.1] - 2022-03-29

### Fixed

- Only list apps from cluster namespace.

## [4.0.0] - 2022-03-29

### Changed

- Update ClusterAPI CR's to `v1beta1`.

## [3.14.1] - 2022-03-11

### Changed

- Update `aws-pod-identity-webhook` app version.

## [3.14.0] - 2022-03-04

### Added

- Add IAM Roles for Service Accounts feature support for AWS.

## [3.13.0] - 2022-01-27

### Changed

- Removed encryption key creation. Encryption keys will be managed by `encryption-provider-operator`.

## [3.12.0] - 2021-12-06

### Changed

- Added support for Azure by selectively disabling features that are AWS specific.

## [3.11.0] - 2021-11-08

### Added

- Check if `kiam-watchdog` app has to be enabled.
- Add cluster CA to cluster values configmap for apps like dex that need to
verify it.

## [3.10.0] - 2021-08-30

### Changed

- Introducing `v1alpha3` CR's.

## [3.9.0] - 2021-07-20

### Changed

- Use `app-operator-konfigure` configmap for the app-operator per workload cluster.

## [3.8.0] - 2021-06-16


### Changed

- Adjust helm chart to be used with `config-controller`.

### Fixed

- Updated OperatorKit to v4.3.1 for Kubernetes 1.20 support.
- Fix `clusterIPRange` value in configmap.
- Fix `kubeconfig` resource to search secrets in all namespaces.

## [3.7.1] - 2021-03-17

### Fixed

- Add `AllowedLabels` to clusterconfigmap resource to prevent unnecessary updates.

## [3.7.0] - 2021-03-15

### Added

- Create app CR for per cluster app-operator instance.
- Add `appfinalizer` resource to remove finalizers from workload cluster app CRs.

## [3.6.1] - 2021-02-23

### Removed

- Do not add `VersionBundle` to new `CertConfig` specs (`CertConfig`s are now versioned using a label). **This change requires using `cert-operator` 1.0.0 or later.**

## [3.6.0] - 2021-02-19

### Fixed

- Fix cluster status computation to correctly display rollbacks, version changes and multiple updates.

### Added

- Add unit tests for cluster status computation

## [3.5.1] - 2021-01-28

### Added

- Check existence of chart tarball for `release` CR `apps` in catalog.

## [3.5.0] - 2021-01-05

### Added

- Add vertical pod autoscaler support.
- Add `appversionlabel` resource to update version labels for optional app CRs.

## [3.4.1] - 2020-12-03

### Fixed

-  Allow annotations from current app CR to remain.

## [3.4.0] - 2020-12-02

### Added

- Add functionality to template `catalog` into `apps` depending on `release` CR.

### Changed

- Update `apiextensions`, `k8sclient`, and `operatorkit` dependencies.
- Update github workflows.

## [3.3.1] - 2020-10-15

### Fixed

- Manage Tenant Cluster API errors gracefully.


## [3.3.0] - 2020-09-28

### Added

- Add etcd client certificates for Prometheus.

### Fixed

- Change pod labels

## [3.2.0] - 2020-09-15

### Added

- Introducing Kubernetes events
- Add monitoring labels.

## [3.1.1] - 2020-08-26

### Fixed

- Fix cluster status is not updated during cluster upgrade

## [3.1.0] - 2020-08-24

### Added

- Add NetworkPolicy.

## [3.0.1] - 2020-08-21

### Fixed

- Fixed condition where reference id does not match with G8sControlplane or MachineDeployment

## [3.0.0] - 2020-08-18

### Changed

- Updated backward incompatible Kubernetes dependencies to v1.18.5.

### Removed

- Remove Helm major version label for chart-operator app CR as it is not used.

## [2.3.2] - 2020-07-31

- Handle error basedomain not found gracefully, so that G8sControlPlane CR and MachineDeployment CRs can be deleted

## [2.3.1] - 2020-07-14

### Fixed

- Fix cluster conditions timestamp which was not set correctly

## [2.3.0] 2020-06-19

### Added

- Add `deletecrs` handler for better CR cleanup.
- Add `controlPlaneStatus` handler for master nodes status.

### Changed

- Remove controller context.
- Bump alpine version to 3.12


## [2.2.0] 2020-05-20

### Added

- Add support for HA Masters certificates.
- Add pod CIDR service implementation using local caching.
- Add Helm major version label to chart-operator app CRs for Helm 3 upgrade.
- Add notes annotation to cluster configmaps to make it clear they should not
be edited by users.

### Changed

- Refactor release access and drop cluster-service dependency.
- Remove pull secret from Helm chart.
- Fetch base domain directly from spec and use local caching.



## [2.1.10] 2020-04-28

### Fixed

- Fix cluster upgrade by fetching release versions in Control Plane controller.



## [2.1.9] 2020-04-23

### Added

- Set `Cluster.Status.InfrastructureReady=true` on common status condition `Created`.

### Changed

- Use release.Revision in Helm chart for Helm 3 support.

### Fixed

- Fix RBAC rules for Control Plane CR reconciliation.



## [2.1.8] 2020-04-17

### Added

- Add Giant Swarm release version to cluster status metrics collector.
- Add Dependabot configuration.

### Changed

- Change resource order for more efficient reconciliation.
- Emit metrics for reconciled runtime objects only.
- Drop CRD management to not ensure CRDs in operators anymore.
- Use release.Revision in Helm chart for Helm 3 support.

### Fixed

- Fix Control Plane CR reconciliation.



## [2.1.7] 2020-04-06

### Fixed

- Fix error handling when creating Tenant Cluster API clients.



## [2.1.6] 2020-04-03

- Switch from dep to Go modules.
- Use architect orb.



## [2.1.5] 2020-03-20

### Added

- First release.



[Unreleased]: https://github.com/giantswarm/cluster-operator/compare/v5.4.0...HEAD
[5.4.0]: https://github.com/giantswarm/cluster-operator/compare/v5.3.0...v5.4.0
[5.3.0]: https://github.com/giantswarm/cluster-operator/compare/v5.2.0...v5.3.0
[5.2.0]: https://github.com/giantswarm/cluster-operator/compare/v5.1.0...v5.2.0
[5.1.0]: https://github.com/giantswarm/cluster-operator/compare/v5.0.0...v5.1.0
[5.0.0]: https://github.com/giantswarm/cluster-operator/compare/v4.6.2...v5.0.0
[4.6.2]: https://github.com/giantswarm/cluster-operator/compare/v4.6.1...v4.6.2
[4.6.1]: https://github.com/giantswarm/cluster-operator/compare/v4.6.0...v4.6.1
[4.6.0]: https://github.com/giantswarm/cluster-operator/compare/v4.5.2...v4.6.0
[4.5.2]: https://github.com/giantswarm/cluster-operator/compare/v4.5.1...v4.5.2
[4.5.1]: https://github.com/giantswarm/cluster-operator/compare/v4.5.0...v4.5.1
[4.5.0]: https://github.com/giantswarm/cluster-operator/compare/v4.4.0...v4.5.0
[4.4.0]: https://github.com/giantswarm/cluster-operator/compare/v4.3.0...v4.4.0
[4.3.0]: https://github.com/giantswarm/cluster-operator/compare/v4.2.0...v4.3.0
[4.2.0]: https://github.com/giantswarm/cluster-operator/compare/v4.1.0...v4.2.0
[4.1.0]: https://github.com/giantswarm/cluster-operator/compare/v4.0.2...v4.1.0
[4.0.2]: https://github.com/giantswarm/cluster-operator/compare/v4.0.1...v4.0.2
[4.0.1]: https://github.com/giantswarm/cluster-operator/compare/v4.0.0...v4.0.1
[4.0.0]: https://github.com/giantswarm/cluster-operator/compare/v3.14.1...v4.0.0
[3.14.1]: https://github.com/giantswarm/cluster-operator/compare/v3.14.0...v3.14.1
[3.14.0]: https://github.com/giantswarm/giantswarm/compare/v3.13.0...v3.14.0
[3.13.0]: https://github.com/giantswarm/cluster-operator/compare/v3.12.0...v3.13.0
[3.12.0]: https://github.com/giantswarm/cluster-operator/compare/v3.11.0...v3.12.0
[3.11.0]: https://github.com/giantswarm/cluster-operator/compare/v3.10.0...v3.11.0
[3.10.0]: https://github.com/giantswarm/cluster-operator/compare/v3.9.0...v3.10.0
[3.9.0]: https://github.com/giantswarm/cluster-operator/compare/v3.8.0...v3.9.0
[3.8.0]: https://github.com/giantswarm/cluster-operator/compare/v3.7.1...v3.8.0
[3.7.1]: https://github.com/giantswarm/cluster-operator/compare/v3.7.0...v3.7.1
[3.7.0]: https://github.com/giantswarm/cluster-operator/compare/v3.6.1...v3.7.0
[3.6.1]: https://github.com/giantswarm/cluster-operator/compare/v3.6.0...v3.6.1
[3.6.0]: https://github.com/giantswarm/cluster-operator/compare/v3.5.1...v3.6.0
[3.5.1]: https://github.com/giantswarm/cluster-operator/compare/v3.5.0...v3.5.1
[3.5.0]: https://github.com/giantswarm/cluster-operator/compare/v3.4.1...v3.5.0
[3.4.1]: https://github.com/giantswarm/cluster-operator/compare/v3.4.0...v3.4.1
[3.4.0]: https://github.com/giantswarm/cluster-operator/compare/v3.3.1...v3.4.0
[3.3.1]: https://github.com/giantswarm/cluster-operator/compare/v3.3.0...v3.3.1
[3.3.0]: https://github.com/giantswarm/cluster-operator/compare/v3.2.0...v3.3.0
[3.2.0]: https://github.com/giantswarm/cluster-operator/compare/v3.1.1...v3.2.0
[3.1.1]: https://github.com/giantswarm/cluster-operator/compare/v3.1.0...v3.1.1
[3.1.0]: https://github.com/giantswarm/cluster-operator/compare/v3.0.1...v3.1.0
[3.0.1]: https://github.com/giantswarm/cluster-operator/compare/v3.0.0...v3.0.1
[3.0.0]: https://github.com/giantswarm/cluster-operator/compare/v2.3.2...v3.0.0
[2.3.2]: https://github.com/giantswarm/cluster-operator/compare/v2.3.1...v2.3.2
[2.3.1]: https://github.com/giantswarm/cluster-operator/compare/v2.3.0...v2.3.1
[2.2.0]: https://github.com/giantswarm/cluster-operator/compare/v2.2.0...v2.3.0
[2.2.0]: https://github.com/giantswarm/cluster-operator/compare/v2.1.10...v2.2.0
[2.1.10]: https://github.com/giantswarm/cluster-operator/compare/v2.1.9...v2.1.10
[2.1.9]: https://github.com/giantswarm/cluster-operator/compare/v2.1.8...v2.1.9
[2.1.8]: https://github.com/giantswarm/cluster-operator/compare/v2.1.7...v2.1.8
[2.1.7]: https://github.com/giantswarm/cluster-operator/compare/v2.1.6...v2.1.7
[2.1.6]: https://github.com/giantswarm/cluster-operator/compare/v2.1.5...v2.1.6

[2.1.5]: https://github.com/giantswarm/cluster-operator/releases/tag/v2.1.5
