package template

// ClusterOperatorChartValues values required by aws-operator-chart, the
// environment variables will be expanded before writing the contents to a file.
const ClusterOperatorChartValues = `Installation:
  V1:
    Guest:
      Kubernetes:
        API:
          ClusterIPRange: 10.0.0.0/16
    Auth:
      Vault:
        Certificate:
          TTL: 3000h
    Secret:
      Registry:
        PullSecret:
          DockerConfigJSON: "{\"auths\":{\"quay.io\":{\"auth\":\"${REGISTRY_PULL_SECRET}\"}}}"
`
