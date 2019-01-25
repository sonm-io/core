package auth

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Event describes fully-qualified gRPC method name.
type Event string

func (e Event) String() string {
	return string(e)
}

// AuthRouter is an entry point of our gRPC authorization.
//
// By default the router allows unregistered events, but this behavior can be
// changed using `DenyUnregistered` option.
type AuthRouter struct {
	mu        sync.RWMutex
	ctx       context.Context
	log       *zap.Logger
	prefix    string
	fallback  Authorization
	verifiers map[Event]Authorization
}

// NewEventAuthorization constructs a new event authorization.
func NewEventAuthorization(ctx context.Context, options ...EventAuthorizationOption) *AuthRouter {
	router := &AuthRouter{
		ctx:       ctx,
		log:       zap.NewNop(),
		verifiers: make(map[Event]Authorization, 0),
		fallback:  NewNilAuthorization(),
	}

	for _, option := range options {
		option(router)
	}

	return router
}

func (r *AuthRouter) addAuthorization(event Event, auth Authorization) {
	r.verifiers[Event(r.prefix+string(event))] = auth
}

func (r *AuthRouter) Authorize(ctx context.Context, event Event, request interface{}) error {
	r.log.Debug("authorizing request", zap.Stringer("method", event))

	return r.AuthorizeNoLog(ctx, event, request)
}

func (r *AuthRouter) AuthorizeNoLog(ctx context.Context, event Event, request interface{}) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	verify, ok := r.verifiers[event]
	if !ok {
		return r.fallback.Authorize(ctx, request)
	}

	return verify.Authorize(ctx, request)
}

// EventAuthorizationOption describes authorization option.
type EventAuthorizationOption func(router *AuthRouter)

// WithLog is an option that assigns the specified logger to be able to log
// some important events.
func WithLog(log *zap.Logger) EventAuthorizationOption {
	return func(router *AuthRouter) {
		router.log = log
	}
}

// WithEventPrefix is an option that specifies event prefix for configuration.
func WithEventPrefix(prefix string) EventAuthorizationOption {
	return func(router *AuthRouter) {
		router.prefix = prefix
	}
}

type AuthOption struct {
	events []string
}

// Allow constructs an AuthOption that is used for further authorization
// attachment.
func Allow(events ...string) AuthOption {
	return AuthOption{
		events: events,
	}
}

// With attaches the given authorization to previously specified events.
func (o AuthOption) With(auth Authorization) EventAuthorizationOption {
	return func(router *AuthRouter) {
		for _, event := range o.events {
			router.addAuthorization(Event(event), auth)
		}
	}
}

// WithFallback constructs an option to assign fallback authorization, that
// will act when an unregistered event comes.
func WithFallback(auth Authorization) EventAuthorizationOption {
	return func(router *AuthRouter) {
		router.fallback = auth
	}
}

type Authorization interface {
	Authorize(ctx context.Context, request interface{}) error
}

type nilAuthorization struct{}

// NewNilAuthorization constructs a new authorization, that will allow any
// incoming event.
func NewNilAuthorization() Authorization {
	return &nilAuthorization{}
}

func (nilAuthorization) Authorize(ctx context.Context, request interface{}) error {
	return nil
}

type denyAuthorization struct{}

// NewDenyAuthorization constructs a new authorization, that will deny any
// incoming event.
func NewDenyAuthorization() Authorization {
	return &denyAuthorization{}
}

func (denyAuthorization) Authorize(ctx context.Context, request interface{}) error {
	return status.Error(codes.Unauthenticated, "permission denied")
}

// NewTransportAuthorization constructs an authorization that allows to call
// methods from the context which has required transport credentials.
// More precisely the caller context must have peer info with verified
// Ethereum address to compare with.
func NewTransportAuthorization(ethAddr common.Address) Authorization {
	return &transportCredentialsAuthorization{
		ethAddr: ethAddr,
	}
}

type transportCredentialsAuthorization struct {
	ethAddr common.Address
}

func (a *transportCredentialsAuthorization) Authorize(ctx context.Context, request interface{}) error {
	peerInfo, err := FromContext(ctx)
	if err != nil {
		return err
	}

	if equalAddresses(peerInfo.Addr, a.ethAddr) {
		return nil
	}

	return status.Errorf(codes.Unauthenticated, "the wallet %s has no access", peerInfo.Addr.Hex())
}

type watchedEntry struct {
	subscribers    []chan<- struct{}
	expirationTime time.Time
}

func (m *watchedEntry) IsExpired() bool {
	return time.Now().After(m.expirationTime)
}

func (m *watchedEntry) AddSubscriber(tx chan<- struct{}) {
	m.subscribers = append(m.subscribers, tx)
}

func (m *watchedEntry) NotifySubscribers() {
	for _, channel := range m.subscribers {
		close(channel)
	}

	m.subscribers = nil
}

// Like "anyOfAuthorization", but allows to dynamically add and remove ETH
// addresses.
type AnyOfTransportCredentialsAuthorization struct {
	mu      sync.RWMutex
	entries map[common.Address]*watchedEntry
}

func NewAnyOfTransportCredentialsAuthorization(ctx context.Context) *AnyOfTransportCredentialsAuthorization {
	m := &AnyOfTransportCredentialsAuthorization{
		entries: map[common.Address]*watchedEntry{},
	}

	go m.run(ctx)

	return m
}

func (m *AnyOfTransportCredentialsAuthorization) run(ctx context.Context) {
	timer := time.NewTicker(time.Second)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			m.checkExpired()
		}
	}
}

func (m *AnyOfTransportCredentialsAuthorization) checkExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for addr, entry := range m.entries {
		if entry.IsExpired() {
			entry.NotifySubscribers()
			delete(m.entries, addr)
		}
	}
}

func (m *AnyOfTransportCredentialsAuthorization) Subscribe(addr common.Address) <-chan struct{} {
	txrx := make(chan struct{})

	m.mu.Lock()
	defer m.mu.Unlock()

	if entry, ok := m.entries[addr]; ok {
		entry.AddSubscriber(txrx)
	} else {
		close(txrx)
	}

	return txrx
}

func (m *AnyOfTransportCredentialsAuthorization) Add(addr common.Address, ttl time.Duration) {
	expirationTime := m.expirationTimeFromDuration(ttl)

	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.entries[addr]
	if ok {
		entry.expirationTime = expirationTime
	} else {
		m.entries[addr] = &watchedEntry{
			expirationTime: expirationTime,
		}
	}
}

func (m *AnyOfTransportCredentialsAuthorization) Remove(addr common.Address) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if entry, ok := m.entries[addr]; ok {
		entry.NotifySubscribers()
	}
	delete(m.entries, addr)
}

func (m *AnyOfTransportCredentialsAuthorization) Authorize(ctx context.Context, request interface{}) error {
	peerInfo, err := FromContext(ctx)
	if err != nil {
		return err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if entry, ok := m.entries[peerInfo.Addr]; ok && !entry.IsExpired() {
		return nil
	}

	return status.Errorf(codes.Unauthenticated, "the wallet %s has no access", peerInfo.Addr.Hex())
}

func (m *AnyOfTransportCredentialsAuthorization) expirationTimeFromDuration(duration time.Duration) time.Time {
	expirationTime := time.Unix(math.MaxInt32, 0)
	if duration != 0 {
		expirationTime = time.Now().Add(duration)
	}

	return expirationTime
}
