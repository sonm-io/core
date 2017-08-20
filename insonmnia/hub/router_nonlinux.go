// +build !linux

package hub

import (
	"context"
	"github.com/sonm-io/core/insonmnia/gateway"
)

func newIPVSRouter(context.Context, *gateway.Gateway) router {
	return newDirectRouter()
}
