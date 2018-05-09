package template

// ClusterOperatorResourceChartValues values required by cluster-operator-resource-chart,
// the environment variables will be expanded before writing the contents to a file.
const ClusterOperatorResourceChartValues = `guest:
  name: "${CLUSTER_NAME}"
  dnsZone: "${CLUSTER_NAME}.${COMMON_DOMAIN_GUEST}"
  id: "${CLUSTER_NAME}"
  owner: "giantswarm"
versionBundle:
  version: "${CLOP_VERSION_BUNDLE_VERSION}"
`
