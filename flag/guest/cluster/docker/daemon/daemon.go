package daemon

// Daemon is a data structure to hold guest cluster Docker daemon specific
// configuration flags.
type Daemon struct {
	CIDR      string
	ExtraArgs string
}
