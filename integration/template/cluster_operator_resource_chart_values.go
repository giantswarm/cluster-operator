package template

// ClusterOperatorResourceChartValues values required by cluster-operator-resource-chart,
// the environment variables will be expanded before writing the contents to a file.
const ClusterOperatorResourceChartValues = `guest:
  name: "${CLUSTER_NAME}"
  dnsZone: "${CLUSTER_NAME}.k8s.${COMMON_DOMAIN}"
  id: "${CLUSTER_NAME}"
  owner: "giantswarm"
  versionBundles:
  - name: aws-operator
    version: ${AWSOP_VERSION_BUNDLE_VERSION}
  - name: cert-operator
    version: ${CERTOP_VERSION_BUNDLE_VERSION}
  - name: cluster-operator
    version: ${CLOP_VERSION_BUNDLE_VERSION}
versionBundle:
  version: "${CLOP_VERSION_BUNDLE_VERSION}"
`
