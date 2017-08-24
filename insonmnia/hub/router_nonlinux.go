// +build !linux

package hub

import (
	"context"
	"github.com/sonm-io/core/insonmnia/gateway"
)

func newIPVSRouter(context.Context, string, *gateway.Gateway, *gateway.PortPool) router {
	return newDirectRouter()
}
