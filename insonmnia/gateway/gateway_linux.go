package gateway

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/tehnerd/gnl2go"

	log "github.com/noxiouz/zapctx/ctxlog"
	"go.uber.org/zap"
)

const (
	PlatformSupportIPVS = true
)

var (
	ErrIPVSFailed      = errors.New("error while calling into IPVS")
	ErrBackendExists   = errors.New("backend already exists")
	ErrServiceNotFound = errors.New("virtual service not found")
	ErrBackendNotFound = errors.New("backend not found")
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

// TODO (3Hren): func (g *Gateway) GetServices() ([]string, error).
// TODO (3Hren): func (g *Gateway) GetBackends(vsID string) ([]string, error).
// TODO (3Hren): func (g *Gateway) GetBackend(vsID, rsID string) (*RealOptions, error).

// CreateService registers a new virtual service with IPVS.
func (g *Gateway) CreateService(vsID string, options *ServiceOptions) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.createService(vsID, options)
}

// CreateBackend registers a new backend with the virtual service.
func (g *Gateway) CreateBackend(vsID, rsID string, options *RealOptions) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.createBackend(vsID, rsID, options)
}

// RemoveService deregisters a virtual service.
func (g *Gateway) RemoveService(vsID string) (*ServiceOptions, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.removeService(vsID)
}

// RemoveBackend deregisters a backend from the virtual service.
func (g *Gateway) RemoveBackend(vsID, rsID string) (*RealOptions, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.removeBackend(vsID, rsID)
}

// Metrics returns some metrics for a given virtual service.
func (g *Gateway) GetMetrics(vsID string) (*Metrics, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.getMetrics(vsID)
}

func (g *Gateway) createService(vsID string, options *ServiceOptions) error {
	log.G(g.ctx).Info("creating virtual service",
		zap.String("service", vsID),
		zap.String("host", options.host.String()),
		zap.Uint16("port", options.Port),
		zap.String("protocol", options.Protocol),
		zap.String("method", options.Method),
	)

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

func (g *Gateway) removeService(vsID string) (*ServiceOptions, error) {
	vs, exists := g.services[vsID]
	if !exists {
		return nil, ErrServiceNotFound
	}

	delete(g.services, vsID)

	log.G(g.ctx).Info("removing virtual service",
		zap.String("service", vsID),
		zap.String("host", vs.options.Host),
		zap.Uint16("port", vs.options.Port),
	)

	if err := g.ipvs.DelService(vs.options.host.String(), vs.options.Port, vs.options.protocol); err != nil {
		log.G(g.ctx).Error("failed to remove virtual service [%s]", zap.String("ID", vsID))
		return nil, ErrIPVSFailed
	}

	for rsID, backend := range g.backends {
		// Filter out non-ours backends.
		if backend.service != vs {
			continue
		}

		log.G(g.ctx).Info("cleaning up now orphaned backend",
			zap.String("service", vsID),
			zap.String("backend", rsID),
		)

		delete(g.backends, rsID)
	}

	return vs.options, nil
}

func (g *Gateway) removeBackend(vsID, rsID string) (*RealOptions, error) {
	rs, exists := g.backends[rsID]
	if !exists {
		return nil, ErrBackendNotFound
	}

	host := rs.service.options.host.String()
	port := rs.service.options.Port
	backendHost := rs.options.host.String()
	backendPort := rs.options.Port
	protocol := rs.service.options.protocol

	log.G(g.ctx).Info("removing backend",
		zap.String("service", vsID),
		zap.String("backend", rsID),
		zap.String("host", host),
		zap.Uint16("port", port),
	)

	if err := g.ipvs.DelDestPort(host, port, backendHost, backendPort, protocol); err != nil {
		log.G(g.ctx).Error("failed to remove backend",
			zap.String("service", vsID),
			zap.String("backend", rsID),
		)
		return nil, ErrIPVSFailed
	}

	delete(g.backends, rsID)

	return rs.options, nil
}

func (g *Gateway) getMetrics(vsID string) (*Metrics, error) {
	vs, exists := g.services[vsID]

	if !exists {
		return nil, ErrServiceNotFound
	}

	stats, err := g.ipvs.GetAllStatsBrief()
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("%s:%s:%d", vs.options.Host, vs.options.Protocol, vs.options.Port)

	for id, stat := range stats {
		if id == key {
			s := stat.GetStats()
			return &Metrics{
				Connections: s["CONNS"],
				InPackets:   s["INPKTS"],
				OutPackets:  s["OUTPKTS"],
				InBytes:     s["INBYTES"],
				OutBytes:    s["OUTBYTES"],

				ConnectionsPerSecond: s["CPS"],
				InPacketsPerSecond:   s["INPPS"],
				OutPacketsPerSecond:  s["OUTPPS"],
				InBytesPerSecond:     s["INBPS"],
				OutBytesPerSecond:    s["OUTBPS"],
			}, nil
		}
	}

	return nil, ErrServiceNotFound
}

func (g *Gateway) Close() {
	log.G(g.ctx).Info("shutting down IPVS context")

	for vsID := range g.services {
		g.RemoveService(vsID)
	}

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
