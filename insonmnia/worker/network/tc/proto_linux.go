package tc

import "golang.org/x/sys/unix"

const (
	ProtoAll Protocol = unix.ETH_P_ALL
	ProtoIP           = unix.ETH_P_IP
)
