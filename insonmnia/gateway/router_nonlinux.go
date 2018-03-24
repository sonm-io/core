// +build !linux

package gateway

import (
	"context"
)

func newIPVSRouter(context.Context, *Gateway, *PortPool) Router {
	return newDirectRouter()
}
