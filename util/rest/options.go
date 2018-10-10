package rest

import (
	"io"
	"net/http"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// options for building rest-server instance
type options struct {
	decoder     Decoder
	encoder     Encoder
	interceptor grpc.UnaryServerInterceptor
	log         *zap.Logger
}

func defaultOptions() *options {
	return &options{
		decoder: &nilDecoder{},
		encoder: &nilEncoder{},
		log:     zap.NewNop(),
	}
}

type nilDecoder struct{}

type nilEncoder struct{}

func (n *nilDecoder) DecodeBody(request *http.Request) (io.Reader, error) {
	return request.Body, nil
}

func (n *nilEncoder) Encode(rw http.ResponseWriter) (http.ResponseWriter, error) {
	return rw, nil
}

// Option func is for applying any params to hub options
type Option func(options *options)

type Decoder interface {
	DecodeBody(request *http.Request) (io.Reader, error)
}

type Encoder interface {
	Encode(rw http.ResponseWriter) (http.ResponseWriter, error)
}

func WithLog(log *zap.Logger) Option {
	return func(o *options) {
		o.log = log
	}
}

func WithDecoder(d Decoder) Option {
	return func(o *options) {
		o.decoder = d
	}
}

func WithEncoder(e Encoder) Option {
	return func(o *options) {
		o.encoder = e
	}
}

func WithInterceptor(interceptor grpc.UnaryServerInterceptor) Option {
	return func(o *options) {
		o.interceptor = interceptor
	}
}

func WithInterceptors(interceptors ...grpc.UnaryServerInterceptor) Option {
	return func(o *options) {
		o.interceptor = grpc_middleware.ChainUnaryServer(interceptors...)
	}
}
