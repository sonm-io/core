package npp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lucas-clemente/quic-go"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/multierror"
	"github.com/sonm-io/core/util/xnet"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

const (
	transportProtocol = "quic"
)

type natPuncherQUICBase struct {
	protocol              string
	rendezvousClient      *rendezvousClientQUIC
	tlsConfig             *tls.Config
	listener              *chanListener
	passiveConnectionTxRx chan connResult
	log                   *zap.SugaredLogger
}

// todo: docs
func newNATPuncherQUICBase(rendezvousClient *rendezvousClientQUIC, tlsConfig *tls.Config, protocol string, log *zap.SugaredLogger) (*natPuncherQUICBase, error) {
	// TODO: Le soutien.
	//  Since we have default "tcp" protocol for TCP punching it is
	//  meaningless for QUIC. Moreover now we treat this parameter as
	//  application protocol.
	if protocol == sonm.DefaultNPPProtocol {
		protocol = "grpc"
	}

	protocol = fmt.Sprintf("%s+%s", transportProtocol, protocol)

	listener, err := quic.Listen(rendezvousClient.PacketConn(), tlsConfig, xnet.DefaultQUICConfig())
	if err != nil {
		return nil, err
	}

	connectionTxRx := make(chan connResult, 64)

	m := &natPuncherQUICBase{
		rendezvousClient:      rendezvousClient,
		tlsConfig:             tlsConfig,
		protocol:              protocol,
		listener:              newChanListener(&xnet.BackPressureListener{Listener: &xnet.QUICListener{Listener: listener}, Log: log.Desugar()}, connectionTxRx),
		passiveConnectionTxRx: connectionTxRx,
		log:                   log.With(zap.String("protocol", protocol)),
	}

	return m, nil
}

func (m *natPuncherQUICBase) RendezvousAddr() net.Addr {
	return m.rendezvousClient.RemoteAddr()
}

func (m *natPuncherQUICBase) punchAddr(ctx context.Context, addr *sonm.Addr) (net.Conn, error) {
	peerAddr, err := addr.IntoUDP()
	if err != nil {
		return nil, err
	}

	udpConn := m.rendezvousClient.PacketConn()

	cfg := xnet.DefaultQUICConfig()

	session, err := quic.DialContext(ctx, udpConn, peerAddr, peerAddr.String(), m.tlsConfig, cfg)
	if err != nil {
		return nil, err
	}

	return xnet.NewQUICConn(session)
}

type natPuncherClientQUIC struct {
	*natPuncherQUICBase
}

func newNATPuncherClientQUIC(rendezvousClient *rendezvousClientQUIC, tlsConfig *tls.Config, protocol string, log *zap.SugaredLogger) (*natPuncherClientQUIC, error) {
	base, err := newNATPuncherQUICBase(rendezvousClient, tlsConfig, protocol, log)
	if err != nil {
		return nil, err
	}

	return &natPuncherClientQUIC{
		natPuncherQUICBase: base,
	}, nil
}

func (m *natPuncherClientQUIC) DialContext(ctx context.Context, addr common.Address) (net.Conn, error) {
	// The first thing we need is to resolve the specified address using Rendezvous server.
	response, err := m.resolve(ctx, addr)
	if err != nil {
		return nil, err
	}

	if response.Empty() {
		return nil, fmt.Errorf("no addresses resolved")
	}

	activeConnectionTxRx := m.punch(ctx, response.GetAddresses())
	defer func() { go drainConnResultChannel(activeConnectionTxRx) }()

	for {
		if activeConnectionTxRx == nil && m.passiveConnectionTxRx == nil {
			return nil, newRendezvousError(fmt.Errorf("failed to dial"))
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case connResult, ok := <-m.passiveConnectionTxRx:
			if !ok {
				m.passiveConnectionTxRx = nil
				continue
			}

			if connResult.Error() != nil {
				m.log.With("punch", "passive").Debugf("received NPP error from: %v", connResult.Error())
				continue
			}

			m.log.With("punch", "passive").Debugf("received NPP connection from %s", connResult.RemoteAddr())
			return connResult.Unwrap()
		case connResult, ok := <-activeConnectionTxRx:
			if !ok {
				activeConnectionTxRx = nil
				continue
			}

			if connResult.Error() != nil {
				m.log.With("punch", "active").Debugf("received NPP error from: %v", connResult.Error())
				continue
			}

			m.log.With("punch", "active").Debugf("received NPP connection from %s", connResult.RemoteAddr())
			return connResult.Unwrap()
		}
	}
}

func (m *natPuncherClientQUIC) Close() error {
	defer func() { go drainConnResultChannel(m.passiveConnectionTxRx) }()

	errs := multierror.NewMultiError()

	if err := m.rendezvousClient.Close(); err != nil {
		errs = multierror.Append(errs, err)
	}

	if err := m.listener.Close(); err != nil {
		errs = multierror.Append(errs, err)
	}

	return errs.ErrorOrNil()
}

func (m *natPuncherClientQUIC) resolve(ctx context.Context, addr common.Address) (*sonm.RendezvousReply, error) {
	request := &sonm.ConnectRequest{
		Protocol:     m.protocol,
		PrivateAddrs: []*sonm.Addr{},
		ID:           addr.Bytes(),
	}

	return m.rendezvousClient.Resolve(ctx, request)
}

func (m *natPuncherClientQUIC) punch(ctx context.Context, addrs []*sonm.Addr) <-chan connResult {
	if len(addrs) == 0 {
		return nil
	}

	channel := make(chan connResult, 1)

	go m.doPunch(ctx, addrs, channel, &clientConnectionWatcher{Log: m.log})

	return channel
}

func (m *natPuncherClientQUIC) doPunch(ctx context.Context, addrs []*sonm.Addr, readinessChannel chan<- connResult, watcher connectionWatcher) {
	defer close(readinessChannel)

	m.log.Debugf("punching %d endpoint(s): %s", len(addrs), sonm.FormatAddrs(addrs...))

	// Pending connection queue. Since we perform all connection attempts
	// asynchronously we must wait until all of them succeeded or errored to
	// prevent both memory and fd leak.
	pendingTxRx := make(chan connResult, len(addrs))
	wg := sync.WaitGroup{}
	wg.Add(len(addrs))

	for _, addr := range addrs {
		addr := addr

		go func() {
			defer wg.Done()
			pendingTxRx <- newConnResult(m.punchAddr(ctx, addr))
		}()
	}

	go func() {
		wg.Wait()
		close(pendingTxRx)
	}()

	var peer net.Conn
	var errs = multierror.NewMultiError()
	for connResult := range pendingTxRx {
		if connResult.Error() != nil {
			m.log.Debugw("received NPP connection candidate notification", zap.Error(connResult.Error()))
			errs = multierror.AppendUnique(errs, connResult.Error())
			continue
		}

		m.log.Debugf("received NPP connection candidate from %s", connResult.RemoteAddr())

		if peer != nil {
			// If we're already established a connection the only thing we can
			// do with the rest - is to put in the queue for further
			// extraction. The client is responsible to close excess
			// connections, while on the our side they will be dropped after
			// being accepted.
			watcher.OnMoreConnections(connResult.conn)
		} else {
			peer = connResult.conn
			// Do not return here - still need to handle possibly successful connections.
			readinessChannel <- newConnResultOk(connResult.conn)
		}
	}

	if peer == nil {
		readinessChannel <- newConnResultErr(fmt.Errorf("failed to punch the network using NPP: all attempts has failed - %s", errs.Error()))
	}
}

type natPuncherServerQUIC struct {
	*natPuncherQUICBase

	readinessTxRx        chan struct{}
	numPunchesInProgress *atomic.Uint32
	activeConnectionTxRx chan connResult
	cancelFunc           context.CancelFunc
}

func newNATPuncherServerQUIC(rendezvousClient *rendezvousClientQUIC, tlsConfig *tls.Config, protocol string, log *zap.SugaredLogger) (*natPuncherServerQUIC, error) {
	base, err := newNATPuncherQUICBase(rendezvousClient, tlsConfig, protocol, log)
	if err != nil {
		return nil, err
	}

	readinessTxRx := make(chan struct{}, 16)

	for i := 0; i < cap(readinessTxRx); i++ {
		readinessTxRx <- struct{}{}
	}

	activeConnectionTxRx := make(chan connResult, 64)

	ctx, cancel := context.WithCancel(context.Background())

	m := &natPuncherServerQUIC{
		natPuncherQUICBase:   base,
		readinessTxRx:        readinessTxRx,
		numPunchesInProgress: atomic.NewUint32(0),
		activeConnectionTxRx: activeConnectionTxRx,
		cancelFunc:           cancel,
	}

	go m.run(ctx)

	return m, nil
}

func (m *natPuncherServerQUIC) AcceptContext(ctx context.Context) (net.Conn, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case connResult := <-m.activeConnectionTxRx:
		if connResult.conn != nil {
			m.log.With("punch", "active").Debugf("received NPP connection from %s", connResult.RemoteAddr())
		}

		return connResult.Unwrap()
	case connResult := <-m.passiveConnectionTxRx:
		if connResult.conn != nil {
			m.log.With("punch", "passive").Debugf("received NPP connection from %s", connResult.RemoteAddr())
		}

		return connResult.Unwrap()
	}
}

func (m *natPuncherServerQUIC) Close() error {
	defer func() { go drainConnResultChannel(m.activeConnectionTxRx) }()
	defer func() { go drainConnResultChannel(m.passiveConnectionTxRx) }()

	m.cancelFunc()

	errs := multierror.NewMultiError()

	if err := m.rendezvousClient.Close(); err != nil {
		errs = multierror.Append(errs, err)
	}

	if err := m.listener.Close(); err != nil {
		errs = multierror.Append(errs, err)
	}

	return errs.ErrorOrNil()
}

func (m *natPuncherServerQUIC) run(ctx context.Context) {
	defer func() {
		defer close(m.activeConnectionTxRx)

		for {
			// No pending punches right now and there won't.
			if m.numPunchesInProgress.Load() == 0 {
				return
			}

			// Otherwise wait for currently processing punches finish.
			<-m.readinessTxRx
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.readinessTxRx:
			m.log.Debugf("publishing on the Rendezvous server")

			publishResponse, err := m.publish(ctx)
			if err != nil {
				m.log.Warnw("failed to publish itself on the rendezvous", zap.Error(err))
				m.activeConnectionTxRx <- newConnResultErr(newRendezvousError(err))
				return
			}

			m.numPunchesInProgress.Inc()
			go m.punch(ctx, []*sonm.Addr{publishResponse.GetPublicAddr()})
		}
	}
}

func (m *natPuncherServerQUIC) publish(ctx context.Context) (*sonm.RendezvousReply, error) {
	request := &sonm.PublishRequest{
		Protocol: m.protocol,
	}

	return m.rendezvousClient.Publish(ctx, request)
}

func (m *natPuncherServerQUIC) punch(ctx context.Context, addrs []*sonm.Addr) {
	defer func() { m.numPunchesInProgress.Dec(); m.readinessTxRx <- struct{}{} }()

	if len(addrs) == 0 {
		return
	}

	readinessChannel := make(chan error, 1)
	go m.doPunch(ctx, addrs, readinessChannel, &serverConnectionWatcher{ConnectionTxRx: m.activeConnectionTxRx, Log: m.log})

	if err := <-readinessChannel; err != nil {
		m.log.Debugf("failed to actively punch %s: %v", sonm.FormatAddrs(addrs...), err)
	}
}

func (m *natPuncherServerQUIC) doPunch(ctx context.Context, addrs []*sonm.Addr, readinessChannel chan<- error, watcher connectionWatcher) {
	m.log.Debugf("punching %d endpoint(s): %s", len(addrs), sonm.FormatAddrs(addrs...))

	// Pending connection queue. Since we perform all connection attempts
	// asynchronously we must wait until all of them succeeded or errored to
	// prevent both memory and fd leak.
	pendingTxRx := make(chan connResult, len(addrs))
	wg := sync.WaitGroup{}
	wg.Add(len(addrs))

	for _, addr := range addrs {
		addr := addr

		go func() {
			defer wg.Done()

			pendingTxRx <- newConnResult(m.punchAddr(ctx, addr))
		}()
	}

	go func() {
		wg.Wait()
		close(pendingTxRx)
	}()

	var peer net.Conn
	var errs = multierror.NewMultiError()
	for connResult := range pendingTxRx {
		if connResult.Error() != nil {
			m.log.Debugw("received NPP connection candidate", zap.Error(connResult.Error()))
			errs = multierror.AppendUnique(errs, connResult.Error())
			continue
		}

		m.log.Debugf("received NPP connection candidate from %s", connResult.RemoteAddr())

		if peer != nil {
			// If we're already established a connection the only thing we can
			// do with the rest - is to put in the queue for further
			// extraction. The client is responsible to close excess
			// connections, while on the our side they will be dropped after
			// being accepted.
			watcher.OnMoreConnections(connResult.conn)
		} else {
			peer = connResult.conn
			m.activeConnectionTxRx <- newConnResultOk(connResult.conn)
			// Do not return here - still need to handle possibly successful connections.
			readinessChannel <- nil
		}
	}

	if peer == nil {
		readinessChannel <- fmt.Errorf("failed to punch the network using NPP: all attempts has failed - %s", errs.Error())
	}
}
