package ssh

import (
	"github.com/sonm-io/core/insonmnia/npp"
)

// ProxyServerConfig specifies SSH proxy server configuration.
type ProxyServerConfig struct {
	Addr string     `yaml:"endpoint" required:"true"`
	NPP  npp.Config `yaml:"npp" required:"true"`
}
