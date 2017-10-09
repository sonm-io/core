package hub

import (
	"context"
	consul "github.com/hashicorp/consul/api"
	consultestutil "github.com/hashicorp/consul/testutil"
	log "github.com/noxiouz/zapctx/ctxlog"
	"strings"
)

type Consul interface {
	KV() *consul.KV
	LockOpts(opts *consul.LockOptions) (*consul.Lock, error)
}

type devConsul struct {
	server *consultestutil.TestServer
	client *consul.Client
}

type consulLogWriter struct {
	ctx    context.Context
	buffer string
}

func (c *consulLogWriter) Write(p []byte) (n int, err error) {
	c.buffer += string(p)
	parts := strings.Split(c.buffer, "\n")
	counter := 0
	for i := 0; i < len(parts)-1; i++ {
		counter += len(parts[i]) + 1
		part := strings.Trim(parts[i], " ")
		subparts := strings.SplitN(part, " ", 4)
		if len(subparts) != 4 || len(subparts[2]) == 0 || subparts[2][0] != '[' {
			log.G(c.ctx).Debug("consul: " + strings.Trim(parts[i], " "))
		} else {
			log.G(c.ctx).Debug("consul: " + strings.Trim(subparts[3], " "))
		}

	}
	c.buffer = c.buffer[counter:]
	return len(p), nil
}

func newDevConsul(ctx context.Context) (*devConsul, error) {
	server, err := consultestutil.NewTestServerConfig(func(c *consultestutil.TestServerConfig) {
		c.Stderr = &consulLogWriter{ctx, ""}
		c.Stdout = &consulLogWriter{ctx, ""}
	})
	if err != nil {
		return nil, err
	}
	go func() {
		defer func() {
			log.G(ctx).Info("consul: stopping embedded server")
			server.Stop()
		}()
		<-ctx.Done()
	}()
	client, err := consul.NewClient(&consul.Config{Address: server.HTTPAddr})
	if err != nil {
		return nil, err
	}
	return &devConsul{server, client}, nil
}

func (d *devConsul) KV() *consul.KV {
	return d.client.KV()
}

func (d *devConsul) LockOpts(opts *consul.LockOptions) (*consul.Lock, error) {
	return d.client.LockOpts(opts)
}
