// +build !linux

package tc

// Actual values doesn't matter for non-linux platforms, because netlink won't
// compile either.
// Use them as a stubs for cross-platform building and testing.
const (
	ProtoAll Protocol = iota
	ProtoIP
)
