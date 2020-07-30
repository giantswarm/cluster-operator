# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.23.14] - 2020-07-30

- Make NGINX optional on KVM, by ignoring existing NGINX IC App CRs which were managed by cluster-operator.

## [0.23.13] - 2020-07-28

### Changed

- Enable NodePort ingress service on KVM.
- Regenerate GitHub workflows, in post-release phase to create a PR instead of trying to push to legacy branch directly.

## [0.23.12] - 2020-07-10

### Changed

- Added GitHub workflows.

[Unreleased]: https://github.com/giantswarm/cluster-operator/compare/v0.23.14...HEAD
[0.23.14]: https://github.com/giantswarm/cluster-operator/compare/v0.23.13...v0.23.14
[0.23.13]: https://github.com/giantswarm/cluster-operator/compare/v0.23.12...v0.23.13
[0.23.12]: https://github.com/giantswarm/cluster-operator/releases/tag/v0.23.12
