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

func (s *Server) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.RequestURI, "/")
	s.log.Debugf("serving URI: %s", r.RequestURI)
	if len(parts) < 3 {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte("invalid uri provided"))
		return
	}
	for i := 3; i < len(parts); i++ {
		if len(parts[i]) != 0 {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte("invalid uri provided"))
			return
		}
	}
	serviceName := parts[1]
	service, ok := s.services[serviceName]
	if !ok {
		s.log.Warnf("service %s not found", serviceName)
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte(fmt.Sprintf("service %s not found", serviceName)))
		return
	}
	methodName := parts[2]
	method, ok := service.methods[methodName]
	if !ok {
		s.log.Warnf("method %s for service %s not found", methodName, serviceName)
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte(fmt.Sprintf("method %s for service %s not found", methodName, serviceName)))
		return
	}

	decodedReader, err := s.decoder.DecodeBody(r)
	if err != nil {
		s.log.Errorf("could not decode body: %s", err)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(fmt.Sprintf("could not decode body: %s", err)))
		return
	}
	rw, err = s.encoder.Encode(rw)
	if err != nil {
		s.log.Errorf("could not encode response writer: %s", err)
		return
	}
	body, _ := ioutil.ReadAll(decodedReader)
	if len(body) == 0 {
		s.log.Error("missing required body")
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(fmt.Sprintf("missing required body")))
		return
	}

	requestValue := reflect.New(method.messageType)
	err = json.Unmarshal(body, requestValue.Interface())
	if err != nil {
		s.log.Errorf("could not unmarshal body: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(fmt.Sprintf("could not unmarshal body: %s", err)))
		return
	}

	h := func(ctx context.Context, req interface{}) (interface{}, error) {
		callParam := []reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(req),
		}

		resp := method.methodValue.Call(callParam)
		if len(resp) != 2 {
			s.log.Errorf("logic error: invalid number of returned arguments: %d", len(resp))
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(fmt.Sprintf("logic error: invalid number of returned arguments")))
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
			Server:     service.service,
			FullMethod: method.fullName,
		}
		result, err = s.interceptor(r.Context(), reflect.Indirect(requestValue).Interface(), info, grpc.UnaryHandler(h))
	} else {
		result, err = h(r.Context(), reflect.Indirect(requestValue).Interface())
	}

	if err != nil {
		rw.WriteHeader(HTTPStatusFromError(err))
		rw.Write([]byte(err.Error()))
	} else {
		data, err := json.Marshal(result)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
		} else {
			rw.Header().Set("content-type", "application/json")
			rw.WriteHeader(http.StatusOK)
			rw.Write(data)
		}
	}
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
