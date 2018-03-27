package rest

import (
	"context"
	"io"
	"net"
	"net/http"

	"google.golang.org/grpc"
)

// options for building hub instance
type options struct {
	ctx         context.Context
	listeners   []net.Listener
	decoder     Decoder
	interceptor grpc.UnaryServerInterceptor
}

func defaultOptions() *options {
	return &options{
		ctx:       context.Background(),
		listeners: []net.Listener{},
		decoder:   &NilDecoder{},
	}
}

type NilDecoder struct{}

func (n *NilDecoder) DecodeBody(request *http.Request) (io.Reader, error) {
	return request.Body, nil
}

// Option func is for applying any params to hub options
type Option func(options *options)

type Decoder interface {
	DecodeBody(request *http.Request) (io.Reader, error)
}

func WithContext(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

func WithListener(l net.Listener) Option {
	return func(o *options) {
		o.listeners = append(o.listeners, l)
	}
}

func WithDecoder(d Decoder) Option {
	return func(o *options) {
		o.decoder = d
	}
}

func WithInterceptor(interceptor grpc.UnaryServerInterceptor) Option {
	return func(o *options) {
		o.interceptor = interceptor
	}
}
