package label

const (
	// CertOperatorVersion sets the version label for CertConfig CRs managed by
	// cert-operator.
	CertOperatorVersion = "cert-operator.giantswarm.io/version"
	// OperatorVersion label transports the operator version requested to be used
	// when reconciling the observed runtime object.
	OperatorVersion = "cluster-operator.giantswarm.io/version"
	// ReleaseVersion is a label specifying a tenant cluster release version.
	ReleaseVersion = "release.giantswarm.io/version"
)
