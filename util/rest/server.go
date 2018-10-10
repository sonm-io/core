package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"strings"

	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type Method struct {
	messageType  reflect.Type
	responseType reflect.Type
	methodValue  reflect.Value
	fullName     string
}

type Service struct {
	methods  map[string]*Method
	service  interface{}
	fullName string
}

type serviceHandle struct {
	Service *Service
	Method  *Method
}

type Server struct {
	servers     []*http.Server
	services    map[string]*Service
	decoder     Decoder
	encoder     Encoder
	interceptor grpc.UnaryServerInterceptor
	log         *zap.SugaredLogger
}

func NewServer(opts ...Option) *Server {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}
	return &Server{
		log:         o.log.Sugar(),
		services:    map[string]*Service{},
		decoder:     o.decoder,
		encoder:     o.encoder,
		interceptor: o.interceptor,
	}
}

func (s *Server) Services() []string {
	var names []string
	for _, service := range s.services {
		names = append(names, service.fullName[1:])
	}

	return names
}

func (s *Server) RegisterService(interfacePtr, concretePtr interface{}) error {
	iface := reflect.TypeOf(interfacePtr).Elem()
	serviceName := iface.Name()
	s.log.Debugf("registering service %s", serviceName)

	service, ok := s.services[serviceName]
	if ok {
		return fmt.Errorf("service %s has been already registered", serviceName)
	}

	fullServiceName := "/" + strings.Replace(iface.String(), "Server", "", 1)

	service = &Service{
		methods:  map[string]*Method{},
		service:  concretePtr,
		fullName: fullServiceName,
	}
	s.services[serviceName] = service

	concrete := reflect.ValueOf(concretePtr)
	if !concrete.Type().Implements(iface) {
		return fmt.Errorf("concrete object of type %s should implement provided interface %s", concrete.Type().Name(), iface.Name())
	}
	for i := 0; i < iface.NumMethod(); i++ {
		method := iface.Method(i)

		//TODO: handle streaming
		if method.Type.NumIn() != 2 {
			s.log.Debugf("skipping streaming for method %s.%s", serviceName, method.Name)
			continue
		}
		messageType := method.Type.In(1)
		sstreamType := reflect.TypeOf((*grpc.ServerStream)(nil)).Elem()
		if messageType.Implements(sstreamType) {
			s.log.Debugf("skipping streaming for method %s.%s", serviceName, method.Name)
			continue
		}
		if method.Type.NumOut() != 2 {
			return fmt.Errorf("could not register method %s for service %s - invalid number of returned arguments", method.Name, serviceName)
		}
		service.methods[method.Name] = &Method{
			messageType:  method.Type.In(1),
			responseType: method.Type.Out(0),
			methodValue:  concrete.MethodByName(method.Name),
			fullName:     fullServiceName + "/" + method.Name,
		}
	}
	return nil
}

func (s *Server) ServeHTTP(rw http.ResponseWriter, request *http.Request) {
	// At first we must do response encoder switching for symmetry.
	//
	// For example when a client with encoding performs invalid request it
	// expects (without providing additional info in headers) an encoded
	// response.
	rw, err := s.encoder.Encode(rw)
	if err != nil {
		s.log.Errorf("could not encode response writer: %s", err)
		return
	}

	code, response := s.serveHTTP(request)
	rw.Header().Set("content-type", "application/json")
	rw.WriteHeader(code)

	switch response := response.(type) {
	case error:
		data, err := json.Marshal(map[string]string{
			"error": response.Error(),
		})
		if err != nil {
			panic("unreachable")
		}

		rw.Write(data)
	case []byte:
		rw.Write(response)
	default:
		rw.Write([]byte(fmt.Sprintf(`{"error": "internal server error: response type is %T"}`, response)))
	}
}

func (s *Server) serveHTTP(request *http.Request) (int, interface{}) {
	s.log.Debugf("serving URI: %s", request.RequestURI)

	methodInfo := xgrpc.ParseMethodInfo(request.RequestURI)
	if methodInfo == nil {
		return http.StatusBadRequest, fmt.Errorf("invalid URI provided")
	}

	serviceHandle, err := s.extractServiceHandle(methodInfo)
	if err != nil {
		return http.StatusNotFound, err
	}

	decodedReader, err := s.decoder.DecodeBody(request)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("could not decode body: %s", err)
	}

	body, _ := ioutil.ReadAll(decodedReader)
	if len(body) == 0 {
		return http.StatusBadRequest, fmt.Errorf("missing required body")
	}

	requestValue := reflect.New(serviceHandle.Method.messageType)
	if err := json.Unmarshal(body, requestValue.Interface()); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("could not unmarshal body: %s", err)
	}

	h := func(ctx context.Context, req interface{}) (interface{}, error) {
		callParam := []reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(req),
		}

		resp := serviceHandle.Method.methodValue.Call(callParam)
		if len(resp) != 2 {
			return nil, errors.New("logic error: invalid number of returned arguments")
		}
		if !resp[1].IsNil() {
			return nil, resp[1].Interface().(error)
		} else {
			return resp[0].Interface(), nil
		}

	}

	var result interface{}
	if s.interceptor != nil {
		info := &grpc.UnaryServerInfo{
			Server:     serviceHandle.Service.service,
			FullMethod: serviceHandle.Service.fullName,
		}
		result, err = s.interceptor(request.Context(), reflect.Indirect(requestValue).Interface(), info, grpc.UnaryHandler(h))
	} else {
		result, err = h(request.Context(), reflect.Indirect(requestValue).Interface())
	}

	if err != nil {
		return HTTPStatusFromError(err), err
	} else {
		data, err := json.Marshal(result)
		if err != nil {
			return http.StatusInternalServerError, err
		} else {
			return http.StatusOK, data
		}
	}
}

func (s *Server) extractServiceHandle(methodInfo *xgrpc.MethodInfo) (*serviceHandle, error) {
	service, ok := s.services[methodInfo.Service]
	if !ok {
		return nil, fmt.Errorf("service %s not found", methodInfo.Service)
	}

	methodInfo.Method = strings.Trim(methodInfo.Method, "/")
	method, ok := service.methods[methodInfo.Method]
	if !ok {
		return nil, fmt.Errorf("method %s for service %s not found", methodInfo.Method, methodInfo.Service)
	}

	return &serviceHandle{
		Service: service,
		Method:  method,
	}, nil
}

func (s *Server) Serve(listeners ...net.Listener) error {
	group := errgroup.Group{}
	for _, lis := range listeners {
		l := lis
		srv := http.Server{}
		srv.Handler = s
		group.Go(func() error {
			s.log.Infof("exposing REST server on %s", l.Addr())
			err := srv.Serve(l)
			s.Close()
			return err
		})
		s.servers = append(s.servers, &srv)
	}
	return group.Wait()
}

func (s *Server) Close() {
	for _, srv := range s.servers {
		srv.Close()
	}
}
