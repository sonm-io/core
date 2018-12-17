package npp

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-reuseport"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/multierror"
	"github.com/sonm-io/core/util/xnet"
	"github.com/uber-go/atomic"
	"go.uber.org/zap"
)

type connectionWatcher interface {
	OnMoreConnections(conn net.Conn)
}

type serverConnectionWatcher struct {
	ConnectionTxRx chan<- connResult
	Log            *zap.SugaredLogger
}

func (m *serverConnectionWatcher) OnMoreConnections(conn net.Conn) {
	m.Log.Debugf("enqueueing pending connection from %s", conn.RemoteAddr().String())
	m.ConnectionTxRx <- newConnResultOk(conn)
}

type clientConnectionWatcher struct {
	Log *zap.SugaredLogger
}

func (m *clientConnectionWatcher) OnMoreConnections(conn net.Conn) {
	addr := conn.RemoteAddr().String()

	m.Log.Debugf("closing pending connection from %s", addr)

	if err := conn.Close(); err != nil {
		m.Log.Warnf("failed to close pending connection from %s: %v", addr, err)
	}
}

// ChanListener wraps the "net.Listener" providing an ability to accept
// connections asynchronously and push them into the specified channel.
//
// It takes the ownership over the given Listener, so it shouldn't be closed.
// However, this listener must be closed explicitly using "Close" method.
// The given channel will be closed when there isn't possible to accept
// connections anymore, i.e. either after critical error or calling "Close".
type chanListener struct {
	listener       net.Listener
	connectionTxRx chan<- connResult
}

// NewChanListener constructs a new channel listener and runs it accepting
// loop.
func newChanListener(listener net.Listener, connectionTxRx chan<- connResult) *chanListener {
	m := &chanListener{
		listener:       listener,
		connectionTxRx: connectionTxRx,
	}

	go m.processEvents()

	return m
}

// ProcessEvents starts processing incoming TCP connections.
func (m *chanListener) processEvents() {
	defer close(m.connectionTxRx)

	for {
		conn, err := m.listener.Accept()
		m.connectionTxRx <- newConnResult(conn, err)

		if err != nil {
			return
		}
	}
}

// PrivateAddrs collects and returns private addresses of a network interfaces
// the listening socket bind on.
// They will be used for connection establishment when both peers are located
// in the same private network.
func (m *chanListener) PrivateAddrs() ([]*sonm.Addr, error) {
	addrs, err := privateAddrs(m.listener.Addr())
	if err != nil {
		return nil, err
	}

	return sonm.TransformNetAddrs(addrs)
}

// Close closes this channel listener by closing the underlying listener.
//
// As a result - the processing loop will be broken and the channel will be
// closed.
func (m *chanListener) Close() error {
	return m.listener.Close()
}

type natPuncherTCPBase struct {
	// Protocol describes the application layer protocol, like "grpc" or "ssh".
	// This will be passed in the "protocol" frame to the Rendezvous server,
	// resulting in something like "tcp+ssh" summary protocol.
	protocol string
	// Listener is the accepting side of the NPP protocol.
	// It is required since the simultaneous TCP connection establishment gives
	// unpredictable results on different platforms, meaning that it is unknown
	// whether the connection arrives to the dialing socket or to the listening
	// one.
	listener *chanListener
	// RendezvousClient is a client to the Rendezvous server.
	// It is managed by this puncher and is closed during executing "Close"
	// method.
	rendezvousClient *rendezvousClient
	// PassiveConnectionTxRx is a channel where all connection results from the
	// local listener will be placed.
	// Errors should be placed here only when everything is bad, i.e. the
	// puncher is unable to continue its work. This can be, for example, when
	// the connection to Rendezvous is dead.
	// Strictly speaking, only in case of non-temporary errors.
	passiveConnectionTxRx chan connResult
	// MaxPunchAttempts shows how many attempts should we made to penetrate the
	// NAT in case of failed connection attempt.
	maxPunchAttempts int
	// Log is a logger used by the puncher for internal logging, mostly debug.
	log *zap.SugaredLogger
}

func newNatPuncherTCPBase(rendezvousClient *rendezvousClient, protocol string, log *zap.SugaredLogger) (*natPuncherTCPBase, error) {
	if len(protocol) == 0 {
		return nil, fmt.Errorf("empty protocol is not allowed")
	}

	// It's important here to reuse the Rendezvous client local address for
	// successful NAT penetration in the case of cone NAT.
	listener, err := reuseport.Listen("tcp", rendezvousClient.LocalAddr().String())
	if err != nil {
		return nil, err
	}

	connectionTxRx := make(chan connResult, 64)

	m := &natPuncherTCPBase{
		protocol:              protocol,
		listener:              newChanListener(&xnet.BackPressureListener{Listener: listener, Log: log.Desugar()}, connectionTxRx),
		rendezvousClient:      rendezvousClient,
		passiveConnectionTxRx: connectionTxRx,
		maxPunchAttempts:      3,
		log:                   log,
	}

	return m, nil
}

// RendezvousAddr returns the remote address of the Rendezvous server that is
// used by this puncher.
func (m *natPuncherTCPBase) RendezvousAddr() net.Addr {
	return m.rendezvousClient.RemoteAddr()
}

// PunchAddr tries to establish a TCP connection to the specified address,
// blocking until either the connection is established or an error occurs.
//
// When the specified context is canceled this method unblocks.
func (m *natPuncherTCPBase) punchAddr(ctx context.Context, addr *sonm.Addr) (net.Conn, error) {
	remoteAddr, err := addr.IntoTCP()
	if err != nil {
		return nil, err
	}

	errs := multierror.NewMultiError()
	for i := 0; i < m.maxPunchAttempts; i++ {
		conn, err := DialContext(ctx, protocol, m.rendezvousClient.LocalAddr().String(), remoteAddr.String())
		if err != nil {
			errs = multierror.AppendUnique(errs, err)
			continue
		}

		return conn, nil
	}

	return nil, errs.ErrorOrNil()
}

type natPuncherCTCP struct {
	*natPuncherTCPBase
}

// NewNATPuncherClientTCP constructs a new client-side NAT puncher over TCP.
//
// The newly created puncher MUST be used exactly ONCE by calling "DialContext"
// and should be closed using "Close" method after getting the result.
func newNATPuncherClientTCP(rendezvousClient *rendezvousClient, protocol string, log *zap.SugaredLogger) (*natPuncherCTCP, error) {
	base, err := newNatPuncherTCPBase(rendezvousClient, protocol, log)
	if err != nil {
		return nil, err
	}

	m := &natPuncherCTCP{
		natPuncherTCPBase: base,
	}

	return m, nil
}

func (m *natPuncherCTCP) DialContext(ctx context.Context, addr common.Address) (net.Conn, error) {
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
				continue
			}

			m.log.With("punch", "active").Debugf("received NPP connection from %s", connResult.RemoteAddr())
			return connResult.Unwrap()
		}
	}
}

func (m *natPuncherCTCP) punch(ctx context.Context, addrs []*sonm.Addr) <-chan connResult {
	if len(addrs) == 0 {
		return nil
	}

	channel := make(chan connResult, 1)

	go m.doPunch(ctx, addrs, channel, &clientConnectionWatcher{Log: m.log})

	return channel
}

func (m *natPuncherCTCP) doPunch(ctx context.Context, addrs []*sonm.Addr, readinessChannel chan<- connResult, watcher connectionWatcher) {
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

func (m *natPuncherCTCP) resolve(ctx context.Context, addr common.Address) (*sonm.RendezvousReply, error) {
	privateAddrs, err := m.listener.PrivateAddrs()
	if err != nil {
		return nil, err
	}

	request := &sonm.ConnectRequest{
		Protocol:     m.protocol,
		PrivateAddrs: privateAddrs,
		ID:           addr.Bytes(),
	}

	return m.rendezvousClient.Resolve(ctx, request)
}

func (m *natPuncherCTCP) Close() error {
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

type natPuncherSTCP struct {
	*natPuncherTCPBase
	// ReadinessTxRx is a channel that performs congestion control for server's
	// announcement.
	readinessTxRx chan struct{}
	// NumPunchesInProgress shows the number currently in progress NAT punching
	// processes.
	numPunchesInProgress *atomic.Uint32
	// ActiveConnectionTxRx is a channel where all actively obtained connection
	// results (i.e. from punching) will be placed.
	activeConnectionTxRx chan connResult
	// CancelFunc is used to cancel internal event loop of this puncher.
	cancelFunc context.CancelFunc
}

// NewNATPuncherServerTCP constructs a new server-side TCP NAT puncher.
//
// The created puncher takes ownership over the specified Rendezvous client
// parameter. However, the puncher must be closed using "Close" method.
//
// Note, that passing empty "protocol" parameter is forbidden and results in
// error.
func newNATPuncherServerTCP(rendezvousClient *rendezvousClient, protocol string, log *zap.SugaredLogger) (*natPuncherSTCP, error) {
	base, err := newNatPuncherTCPBase(rendezvousClient, protocol, log)
	if err != nil {
		return nil, err
	}

	readinessTxRx := make(chan struct{}, 16)

	for i := 0; i < cap(readinessTxRx); i++ {
		readinessTxRx <- struct{}{}
	}

	activeConnectionTxRx := make(chan connResult, 64)

	ctx, cancel := context.WithCancel(context.Background())

	m := &natPuncherSTCP{
		natPuncherTCPBase:    base,
		readinessTxRx:        readinessTxRx,
		numPunchesInProgress: atomic.NewUint32(0),
		activeConnectionTxRx: activeConnectionTxRx,
		cancelFunc:           cancel,
	}

	go m.run(ctx)

	return m, nil
}

// AcceptContext blocks the current execution context until a new connection
// arrives or the specified context is canceled.
//
// If the returned error does not implements "Temporary" interface, then this
// puncher must not be used anymore and should be explicitly closed by
// executing "Close" method.
func (m *natPuncherSTCP) AcceptContext(ctx context.Context) (net.Conn, error) {
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

func (m *natPuncherSTCP) run(ctx context.Context) {
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
			go m.punch(ctx, publishResponse.GetAddresses())
		}
	}
}

// Publish announces this server on the Rendezvous, blocking until a remote
// peer expresses the desire to establish a TCP connection, possibly punching
// NATs between.
// When the specified context is canceled this method unblocks.
func (m *natPuncherSTCP) publish(ctx context.Context) (*sonm.RendezvousReply, error) {
	privateAddrs, err := m.listener.PrivateAddrs()
	if err != nil {
		return nil, err
	}

	request := &sonm.PublishRequest{
		Protocol:     m.protocol,
		PrivateAddrs: privateAddrs,
	}

	return m.rendezvousClient.Publish(ctx, request)
}

func (m *natPuncherSTCP) punch(ctx context.Context, addrs []*sonm.Addr) {
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

func (m *natPuncherSTCP) doPunch(ctx context.Context, addrs []*sonm.Addr, readinessChannel chan<- error, watcher connectionWatcher) {
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

// Close closes this TCP NAT puncher, freeing all associated resources.
//
// The puncher becomes unusable after calling this method.
func (m *natPuncherSTCP) Close() error {
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
