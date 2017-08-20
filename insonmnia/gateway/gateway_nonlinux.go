// +build !linux

package gateway

import "context"

const (
	PlatformSupportIPVS = false
)

type Gateway struct{}

func NewGateway(context.Context) (*Gateway, error) {
	return &Gateway{}, nil
}

func (g *Gateway) Close() {}
