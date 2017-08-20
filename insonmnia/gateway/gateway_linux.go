package gateway

import (
	"context"
	"errors"
	"net"

	"github.com/tehnerd/gnl2go"

	log "github.com/noxiouz/zapctx/ctxlog"
	"go.uber.org/zap"
	"sync"
)

const (
	PlatformSupportIPVS = true
)

var (
	ErrIPVSFailed      = errors.New("error while calling into IPVS")
	ErrBackendExists   = errors.New("backend already exists")
	ErrServiceNotFound = errors.New("virtual service not found")
)

type service struct {
	options *ServiceOptions
}

type backend struct {
	options *RealOptions
	service *service
}

type Gateway struct {
	ctx      context.Context
	ipvs     gnl2go.IpvsClient
	endpoint net.IP
	services map[string]*service
	backends map[string]*backend

	mu sync.Mutex
}

func NewGateway(ctx context.Context) (*Gateway, error) {
	log.G(ctx).Info("initializing IPVS context")

	gateway := &Gateway{
		ctx:      ctx,
		ipvs:     gnl2go.IpvsClient{},
		services: make(map[string]*service),
		backends: make(map[string]*backend),
	}

	if err := gateway.ipvs.Init(); err != nil {
		gateway.Close()
		return nil, ErrIPVSFailed
	}

	if gateway.ipvs.Flush() != nil {
		log.G(ctx).Error("failed to clean up IPVS pools - ensure `ip_vs` is loaded")
		gateway.Close()
		return nil, ErrIPVSFailed
	}

	return gateway, nil
}

// CreateService registers a new virtual service with IPVS.
func (g *Gateway) CreateService(vsID string, options *ServiceOptions) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.createService(vsID, options)
}

func (g *Gateway) CreateBackend(vsID, rsID string, options *RealOptions) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.createBackend(vsID, rsID, options)
}

func (g *Gateway) createService(vsID string, options *ServiceOptions) error {
	if err := g.ipvs.AddService(options.host.String(), options.Port, options.protocol, options.Method); err != nil {
		log.G(g.ctx).Error("failed to create creating virtual service", zap.Error(err))
		return err
	}

	g.services[vsID] = &service{options: options}

	return nil
}

func (g *Gateway) createBackend(vsID, rsID string, options *RealOptions) error {
	if _, exists := g.backends[rsID]; exists {
		return ErrBackendExists
	}

	vs, exists := g.services[vsID]

	if !exists {
		return ErrServiceNotFound
	}

	// TODO: Check for AF.

	log.G(g.ctx).Info("creating backend for virtual service",
		zap.String("backend", rsID),
		zap.String("host", options.Host),
		zap.Uint16("port", options.Port),
		zap.String("service", vsID),
	)

	if err := g.ipvs.AddDestPort(vs.options.host.String(), vs.options.Port, options.host.String(), options.Port,
		vs.options.protocol, int32(options.Weight), options.methodID); err != nil {
		log.G(g.ctx).Error("failed to create backend", zap.Error(err))
		return ErrIPVSFailed
	}

	g.backends[rsID] = &backend{options: options, service: vs}

	return nil
}

func (g *Gateway) Close() {
	log.G(g.ctx).Info("shutting down IPVS context")

	//for vsID := range ctx.services {
	//	ctx.RemoveService(vsID)
	//}

	g.ipvs.Exit()
}

func GetOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}
