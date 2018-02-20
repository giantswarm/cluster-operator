package certconfig

// Key is an interface that defines contract between Key implementations in CRD
// frameworks and certconfig resource.
type Key interface {
	KeyError
}
