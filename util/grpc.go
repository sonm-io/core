package util

import (
	"strings"
)

const (
	WorkerAddressHeader = "x_worker_eth_addr"
)

func ExtractMethod(fullMethod string) string {
	parts := strings.Split(fullMethod, "/")
	return parts[len(parts)-1]
}
