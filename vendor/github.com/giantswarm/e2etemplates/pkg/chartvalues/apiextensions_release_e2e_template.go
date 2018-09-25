package chartvalues

const apiExtensionsReleaseE2ETemplate = `namespace: {{ .Namespace }}
operator:
  name: {{ .Operator.Name }}
  version: {{ .Operator.Version }}
versionBundle:
  version: {{ .VersionBundle.Version }}
`
