package docker

// Docker is a data structure to hold guest cluster Docker specific
// configuration flags.
type Docker struct {
	Daemon Daemon
}

// Daemon is a data structure to hold guest cluster Docker daemon specific
// configuration flags.
type Daemon struct {
	CIDR      string
	ExtraArgs string
}
