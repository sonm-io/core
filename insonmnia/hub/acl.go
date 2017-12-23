package hub

import (
	"context"
	"encoding/json"
	"reflect"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"gopkg.in/fatih/set.v0"

	log "github.com/noxiouz/zapctx/ctxlog"
)

var (
	errNoPeerInfo       = status.Error(codes.Unauthenticated, "no peer info")
	errNoDealProvided   = status.Error(codes.Unauthenticated, "no `deal` metadata provided")
	errNoDealFieldFound = status.Error(codes.Internal, "no `Deal` field found")
	errInvalidDealField = status.Error(codes.Internal, "invalid `Deal` field type")
	errNoTaskFieldFound = status.Errorf(codes.Internal, "no task `ID` field found")
	errInvalidTaskField = status.Error(codes.Internal, "invalid task `ID` field type")
	errNoMetadata       = status.Error(codes.Unauthenticated, "no metadata provided")
	errNoWalletProvided = status.Error(codes.Unauthenticated, "no wallet provided")
)

// ACLStorage describes an ACL storage for workers.
//
// A worker connection can be accepted only and the only if its credentials
// provided with the certificate contains in this storage.
type ACLStorage interface {
	// Insert inserts the given worker credentials to the storage.
	Insert(credentials string)
	// Remove removes the given worker credentials from the storage.
	// Returns true if it was actually removed.
	Remove(credentials string) bool
	// Has checks whether the given worker credentials contains in the
	// storage.
	Has(credentials string) bool
	// Each applies the specified function to each credentials in the storage.
	// Traversal will continue until all items in the Set have been visited,
	// or if the closure returns false.
	Each(fn func(string) bool)
}

type workerACLStorage struct {
	storage *set.SetNonTS
	mu      sync.RWMutex
}

func NewACLStorage() ACLStorage {
	return &workerACLStorage{
		storage: set.NewNonTS(),
	}
}

func (s *workerACLStorage) MarshalJSON() ([]byte, error) {
	if s == nil {
		return json.Marshal(nil)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	set := make([]string, 0)
	s.storage.Each(func(item interface{}) bool {
		set = append(set, item.(string))
		return true
	})
	return json.Marshal(set)
}

func (s *workerACLStorage) UnmarshalJSON(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	unmarshalled := make([]string, 0)
	err := json.Unmarshal(data, &unmarshalled)
	if err != nil {
		return err
	}
	s.storage = set.NewNonTS()

	for _, val := range unmarshalled {
		s.storage.Add(val)
	}
	return nil
}

func (s *workerACLStorage) Insert(credentials string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.storage.Add(credentials)
}

func (s *workerACLStorage) Remove(credentials string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	exists := s.storage.Has(credentials)
	if exists {
		s.storage.Remove(credentials)
	}
	return exists
}

func (s *workerACLStorage) Has(credentials string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.storage.Has(credentials)
}

func (s *workerACLStorage) Each(fn func(string) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.storage.Each(func(credentials interface{}) bool {
		return fn(credentials.(string))
	})
}

// Method describes GRPC event, i.e some method name.
type method string

type eventACL struct {
	ctx       context.Context
	mu        sync.RWMutex
	verifiers map[method]EventAuthorization
}

func newEventACL(ctx context.Context) *eventACL {
	return &eventACL{
		ctx:       ctx,
		verifiers: make(map[method]EventAuthorization, 0),
	}
}

func (e *eventACL) authorize(ctx context.Context, method method, request interface{}) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	log.G(e.ctx).Debug("authorizing request", zap.String("method", string(method)))

	authorization, ok := e.verifiers[method]
	if !ok {
		return nil
	}

	return authorization.Authorize(ctx, request)
}

func (e *eventACL) addAuthorization(method method, auth EventAuthorization) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.verifiers[method] = auth
}

type EventAuthorization interface {
	Authorize(ctx context.Context, request interface{}) error
}

type hubManagementAuthorization struct {
	ctx     context.Context
	ethAddr common.Address
}

func newHubManagementAuthorization(ctx context.Context, ethAddr common.Address) EventAuthorization {
	return &hubManagementAuthorization{
		ctx:     ctx,
		ethAddr: ethAddr,
	}
}

func (h *hubManagementAuthorization) Authorize(ctx context.Context, request interface{}) error {
	peerInfo, ok := peer.FromContext(ctx)
	if !ok {
		return errNoPeerInfo
	}

	switch auth := peerInfo.AuthInfo.(type) {
	case util.EthAuthInfo:
		if util.EqualAddresses(auth.Wallet, h.ethAddr) {
			return nil
		}
		return status.Errorf(codes.Unauthenticated, "the wallet %s has no access", auth.Wallet)
	default:
		return status.Error(codes.Unauthenticated, "unknown auth info")
	}
}

// DealMetaData allows to extract deal-specific parameters for
// authorization.
// We implement this interface for all methods that require wallet
// authorization.
type DealMetaData interface {
	// Deal extracts deal ID from the request.
	Deal(ctx context.Context, request interface{}) (DealID, error)
	// Wallet extracts self-signed wallet from the request.
	Wallet(ctx context.Context, request interface{}) (string, error)
}

type dealAuthorization struct {
	ctx      context.Context
	hub      *Hub
	metaData DealMetaData
}

func newDealAuthorization(ctx context.Context, hub *Hub, metaData DealMetaData) EventAuthorization {
	return &dealAuthorization{
		ctx:      ctx,
		hub:      hub,
		metaData: metaData,
	}
}

func (d *dealAuthorization) Authorize(ctx context.Context, request interface{}) error {
	dealID, err := d.metaData.Deal(ctx, request)
	if err != nil {
		return err
	}

	signedWallet, err := d.metaData.Wallet(ctx, request)
	if err != nil {
		return err
	}

	meta, err := d.hub.getDealMeta(dealID)

	if err != nil {
		return err
	}

	allowedWallet := meta.Order.GetByuerID()

	log.G(d.ctx).Debug("found allowed wallet for a deal",
		zap.String("deal", string(dealID)),
		zap.String("signedWallet", signedWallet),
		zap.String("allowedWallet", allowedWallet),
	)

	recoveredAddr, err := util.VerifySelfSignedWallet(signedWallet)
	if err != nil {
		return err
	}

	log.G(d.ctx).Debug("recovered address", zap.String("addr", string(recoveredAddr)))

	if allowedWallet != recoveredAddr {
		return status.Errorf(codes.Unauthenticated, "wallet mismatch: %s", recoveredAddr)
	}

	return nil
}

func extractWalletFromContext(ctx context.Context) (string, error) {
	peerInfo, ok := peer.FromContext(ctx)
	if !ok {
		return "", errNoPeerInfo
	}

	switch peerInfo.AuthInfo.(type) {
	case util.EthAuthInfo:
		md, ok := metadata.FromContext(ctx)
		if !ok {
			return "", errNoMetadata
		}

		walletMD := md["wallet"]
		if len(walletMD) == 0 {
			return "", errNoWalletProvided
		}

		return walletMD[0], nil
	default:
		return "", status.Errorf(codes.Unauthenticated, "unknown auth info %T", peerInfo.AuthInfo)
	}
}

// FieldDealMetaData is a deal meta info extractor that requires the specified
// request to have "sonm.Deal" field and "wallet" metadata.
type fieldDealMetaData struct{}

func (fieldDealMetaData) Deal(ctx context.Context, request interface{}) (DealID, error) {
	requestValue := reflect.Indirect(reflect.ValueOf(request))
	deal := reflect.Indirect(requestValue.FieldByName("Deal"))
	if !deal.IsValid() {
		return "", errNoDealFieldFound
	}

	if deal.Type().Kind() != reflect.Struct {
		return "", errInvalidDealField
	}

	dealId := reflect.Indirect(deal).FieldByName("Id")
	if !dealId.IsValid() {
		return "", errInvalidDealField
	}

	if dealId.Type().Kind() != reflect.String {
		return "", errInvalidDealField
	}

	return DealID(dealId.String()), nil
}

func (fieldDealMetaData) Wallet(ctx context.Context, request interface{}) (string, error) {
	return extractWalletFromContext(ctx)
}

// ContextDealMetaData is a deal meta info extractor that requires the
// specified context to have both "deal" and "wallet" metadata.
type contextDealMetaData struct{}

func (contextDealMetaData) Deal(ctx context.Context, request interface{}) (DealID, error) {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return "", errNoPeerInfo
	}

	dealMD := md["deal"]
	if len(dealMD) == 0 {
		return "", errNoDealProvided
	}

	return DealID(dealMD[0]), nil
}

func (contextDealMetaData) Wallet(ctx context.Context, request interface{}) (string, error) {
	return extractWalletFromContext(ctx)
}

// TaskFieldDealMetaData is a deal meta info extractor that requires the
// specified request to have "Id" field, which is task id, and the context to
// have "wallet" metadata.
type taskFieldDealMetaData struct {
	hub *Hub
}

func (t *taskFieldDealMetaData) Deal(ctx context.Context, request interface{}) (DealID, error) {
	requestValue := reflect.Indirect(reflect.ValueOf(request))
	taskID := reflect.Indirect(requestValue.FieldByName("Id"))
	if !taskID.IsValid() {
		return "", errNoTaskFieldFound
	}

	if taskID.Type().Kind() != reflect.String {
		return "", errInvalidTaskField
	}

	taskInfo, ok := t.hub.tasks[taskID.String()]
	if !ok {
		return "", errTaskNotFound
	}

	return DealID(taskInfo.GetDealId()), nil
}

func (taskFieldDealMetaData) Wallet(ctx context.Context, request interface{}) (string, error) {
	return extractWalletFromContext(ctx)
}
