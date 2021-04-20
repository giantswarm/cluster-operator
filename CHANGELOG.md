# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Dropped ensuring cluster CRDs from controllers.

## [0.27.0] - 2021-04-15

### Changed

- Adjust helm chart to be used with `config-controller`.

## [0.26.0] - 2021-03-25

### Added

- Assign app catalog name from the component in release CR.

## [0.25.1] - 2021-03-17

### Fixed

- Add `AllowedLabels` to clusterconfigmap resource to prevent unnecessary updates.

## [0.25.0] - 2021-03-15

### Added

- Create app CR for per cluster app-operator instance.
- Add `appfinalizer` resource to remove finalizers from workload cluster app CRs.

## [0.24.2] - 2021-02-25

### Changed

- Migrate to Go modules.
- Update `certs` package to v2.0.0.
- Refactor to use slightly newer dependency versions.

## [0.24.1] - 2021-02-23

### Changed

- Align version bundle version and project version.

## [0.24.0] - 2021-02-23

### Changed

- Remove `VersionBundle` version from `CertConfigs` and add the `cert-operator.giantswarm.io/version` label. **This change requires using `cert-operator` 1.0.0 or later**.

## [0.23.22] - 2021-01-29

### Changed

Replacement for [0.23.21] because of incorrect `bundleVersion` in [0.23.21]

## [0.23.21] - 2021-01-28

### Added

- Check existence of chart tarball for `release` CR `apps` in catalog.

## [0.23.20] - 2021-01-07

### Added

- Add appversionlabel resource to update version labels for optional app CRs.

## [0.23.19] - 2020-12-03

### Fixed

-  Allow annotations from current app CR to remain.

## [0.23.18] - 2020-10-21

## [0.23.17] - 2020-10-19

### Changed

- Delete all chartconfig migration logic.

## [0.23.16] - 2020-08-18

### Changed

- Get app-operator version from releases CR.

## [0.23.14] - 2020-07-30

- Make NGINX optional on KVM, by ignoring existing NGINX IC App CRs which were managed by cluster-operator.

## [0.23.13] - 2020-07-28

### Changed

- Enable NodePort ingress service on KVM.
- Regenerate GitHub workflows, in post-release phase to create a PR instead of trying to push to legacy branch directly.

## [0.23.12] - 2020-07-10

### Changed

- Added GitHub workflows.

[Unreleased]: https://github.com/giantswarm/cluster-operator/compare/v0.27.0...HEAD
[0.27.0]: https://github.com/giantswarm/cluster-operator/compare/v0.26.0...v0.27.0
[0.26.0]: https://github.com/giantswarm/cluster-operator/compare/v0.25.1...v0.26.0
[0.25.1]: https://github.com/giantswarm/cluster-operator/compare/v0.25.0...v0.25.1
[0.25.0]: https://github.com/giantswarm/cluster-operator/compare/v0.24.2...v0.25.0
[0.24.2]: https://github.com/giantswarm/cluster-operator/compare/v0.24.1...v0.24.2
[0.24.1]: https://github.com/giantswarm/cluster-operator/compare/v0.24.0...v0.24.1
[0.24.0]: https://github.com/giantswarm/cluster-operator/compare/v0.23.22...v0.24.0
[0.23.22]: https://github.com/giantswarm/cluster-operator/compare/v0.23.21...v0.23.22
[0.23.21]: https://github.com/giantswarm/cluster-operator/compare/v0.23.19...v0.23.21
[0.23.19]: https://github.com/giantswarm/cluster-operator/compare/v0.23.18...v0.23.19
[0.23.18]: https://github.com/giantswarm/cluster-operator/compare/v0.23.17...v0.23.18
[0.23.17]: https://github.com/giantswarm/cluster-operator/compare/v0.23.16...v0.23.17
[0.23.16]: https://github.com/giantswarm/cluster-operator/compare/v0.23.15...v0.23.16
[0.23.15]: https://github.com/giantswarm/cluster-operator/compare/v0.23.14...v0.23.15
[0.23.14]: https://github.com/giantswarm/cluster-operator/compare/v0.23.13...v0.23.14
[0.23.13]: https://github.com/giantswarm/cluster-operator/compare/v0.23.12...v0.23.13
[0.23.12]: https://github.com/giantswarm/cluster-operator/releases/tag/v0.23.12
