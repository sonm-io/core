package hub

import (
	"bytes"
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
	buffer bytes.Buffer
}

func (c *consulLogWriter) Write(p []byte) (n int, err error) {
	c.buffer.Write(p)
	for {
		line, err := c.buffer.ReadBytes('\n')
		if err != nil {
			for i := 0; i < len(line); i++ {
				c.buffer.UnreadByte()
			}
			break
		}
		part := strings.Trim(string(line), " \n")
		subparts := strings.SplitN(part, " ", 4)
		if len(subparts) != 4 || len(subparts[2]) == 0 || subparts[2][0] != '[' {
			log.G(c.ctx).Debug("consul: " + strings.Trim(string(line), " "))
		} else {
			log.G(c.ctx).Debug("consul: " + strings.Trim(subparts[3], " "))
		}
	}
	return len(p), nil
}

func newDevConsul(ctx context.Context) (*devConsul, error) {
	server, err := consultestutil.NewTestServerConfig(func(c *consultestutil.TestServerConfig) {
		c.Stderr = &consulLogWriter{ctx, bytes.Buffer{}}
		c.Stdout = &consulLogWriter{ctx, bytes.Buffer{}}
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
