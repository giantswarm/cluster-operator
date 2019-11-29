[![CircleCI](https://circleci.com/gh/giantswarm/cluster-operator.svg?&style=shield&circle-token=373dcae33aecb47a0a53c51105e9381dff5b0b88)](https://circleci.com/gh/giantswarm/cluster-operator) [![Docker Repository on Quay](https://quay.io/repository/giantswarm/cluster-operator/status "Docker Repository on Quay")](https://quay.io/repository/giantswarm/cluster-operator)

# cluster-operator

The cluster-operator is an in-cluster agent that handles Giant Swarm guest
cluster specific resources.

## Branches

- `thiccc`
    - Up to and including version v0.21.0.
    - Contains all versions of legacy controllers (reconciling
      {AWS,Azure,KVM}ClusterConfig CRs) up to and including v0.21.0.
- `legacy`
    - From version v0.21.1 up to and including v0.x.x.
    - Contains only the latest version of legacy controllers (reconciling
      {AWS,Azure,KVM}ClusterConfig CRs).
- `master`
    - From version v2.0.0.
    - Contains only the latest version of controllers (reconciling cluster API
      objects).

## Getting Project

Clone the git repository: https://github.com/giantswarm/cluster-operator.git

### How to build

Build it using the standard `go build` command.

```
go build github.com/giantswarm/cluster-operator
```

## Contact

- Mailing list: [giantswarm](https://groups.google.com/forum/!forum/giantswarm)
- IRC: #[giantswarm](irc://irc.freenode.org:6667/#giantswarm) on freenode.org
- Bugs: [issues](https://github.com/giantswarm/cluster-operator/issues)

## Contributing & Reporting Bugs

See [CONTRIBUTING](CONTRIBUTING.md) for details on submitting patches, the
contribution workflow as well as reporting bugs.



## License

cluster-operator is under the Apache 2.0 license. See the [LICENSE](LICENSE) file for
details.

