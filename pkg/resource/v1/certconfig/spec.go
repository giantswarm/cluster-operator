package certconfig

// Key is an interface that defines contract between Key implementations in CRD
// frameworks and certconfig resource.
type Key interface {
	KeyError
}

// KeyError is an interface that defines contract between Key error
// check implementations in CRD frameworks and certconfig resource.
type KeyError interface {
	// IsWrongTypeError asserts if error is caused by type mismatch.
	IsWrongTypeError(err error) bool
}
