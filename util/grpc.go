package util

import (
	"strings"
)

const (
	WorkerAddressHeader = "x-worker-eth-addr"
)

func ExtractMethod(fullMethod string) string {
	parts := strings.Split(fullMethod, "/")
	return parts[len(parts)-1]
}
