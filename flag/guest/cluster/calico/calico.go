package calico

// Calico is a data structure to hold guest cluster Calico specific
// configuration flags.
type Calico struct {
	CIDR   int
	MTU    int
	Subnet string
}
