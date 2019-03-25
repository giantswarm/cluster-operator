package kubeconfig

const (
	kubeconfigYaml = `apiVersion: v1
kind: Config
clusters:
- name: giantswarm-w7utg
  cluster:
    server: api.giantswarm.io
    certificate-authority-data: Y2E=
users:
- name: giantswarm-w7utg-user
  user:
    client-certificate-data: Y3J0
    client-key-data: a2V5
contexts:
- name: giantswarm-w7utg-context
  context:
    cluster: giantswarm-w7utg
    user: giantswarm-w7utg-user
current-context: giantswarm-w7utg-context
preferences: {}
`
)
