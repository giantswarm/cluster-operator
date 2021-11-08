# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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



[Unreleased]: https://github.com/giantswarm/cluster-operator/compare/v3.11.0...HEAD
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
