package auth

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
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
	r.mu.RLock()
	defer r.mu.RUnlock()

	r.log.Debug("authorizing request", zap.Stringer("method", event))

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
	peerInfo, ok := peer.FromContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "no peer info")
	}

	switch authInfo := peerInfo.AuthInfo.(type) {
	case EthAuthInfo:
		if equalAddresses(authInfo.Wallet, a.ethAddr) {
			return nil
		}
		return status.Errorf(codes.Unauthenticated, "the wallet %s has no access", authInfo.Wallet.Hex())
	default:
		return status.Error(codes.Unauthenticated, "unknown auth info")
	}
}
