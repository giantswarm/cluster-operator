# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).



## [Unreleased]

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



[Unreleased]: https://github.com/giantswarm/cluster-operator/compare/v2.3.2...HEAD
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
