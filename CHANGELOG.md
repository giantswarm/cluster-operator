# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/giantswarm/cluster-operator/compare/v0.23.22...HEAD
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
