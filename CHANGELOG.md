# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).



## [Unreleased]

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



[Unreleased]: https://github.com/giantswarm/cluster-operator/compare/v2.1.8...HEAD

[2.1.8]: https://github.com/giantswarm/cluster-operator/compare/v2.1.7...v2.1.8
[2.1.7]: https://github.com/giantswarm/cluster-operator/compare/v2.1.6...v2.1.7
[2.1.6]: https://github.com/giantswarm/cluster-operator/compare/v2.1.5...v2.1.6

[2.1.5]: https://github.com/giantswarm/cluster-operator/releases/tag/v2.1.5
