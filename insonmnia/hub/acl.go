package hub

import (
	"context"
	"encoding/json"
	"reflect"
	"sync"

	"github.com/sonm-io/core/insonmnia/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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

// DealMetaData allows to extract deal-specific parameters for
// authorization.
// We implement this interface for all methods that require wallet
// authorization.
type DealMetaData interface {
	// Deal extracts deal ID from the request.
	Deal(ctx context.Context, request interface{}) (DealID, error)
}

type dealAuthorization struct {
	ctx      context.Context
	hub      *Hub
	metaData DealMetaData
}

func newDealAuthorization(ctx context.Context, hub *Hub, metaData DealMetaData) auth.Authorization {
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

	wallet, err := auth.ExtractWalletFromContext(ctx)
	if err != nil {
		return err
	}

	peerWallet := wallet.Hex()
	meta, err := d.hub.getDealMeta(dealID)

	if err != nil {
		return err
	}

	allowedWallet := meta.Order.GetByuerID()

	log.G(d.ctx).Debug("found allowed wallet for a deal",
		zap.Stringer("deal", dealID),
		zap.String("wallet", peerWallet),
		zap.String("allowedWallet", allowedWallet),
	)

	if allowedWallet != peerWallet {
		return status.Errorf(codes.Unauthenticated, "wallet mismatch: %s", peerWallet)
	}

	return nil
}

// FieldDealMetaData is a deal meta info extractor that requires the specified
// request to have "sonm.Deal" field.
type fieldDealMetaData struct{}

func newFieldDealMetaData() DealMetaData {
	return &fieldDealMetaData{}
}

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

// ContextDealMetaData is a deal meta info extractor that requires the
// specified context to have "deal" metadata.
type contextDealMetaData struct{}

func newContextDealMetaData() DealMetaData {
	return &contextDealMetaData{}
}

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

// FieldIdMetaData is a deal meta info extractor that requires the specified
// request to have "Id" field, which is the task id.
type fieldIdMetaData struct {
	hub *Hub
}

func newFieldIdMetaData(hub *Hub) DealMetaData {
	return &fieldIdMetaData{hub: hub}
}

func (t *fieldIdMetaData) Deal(ctx context.Context, request interface{}) (DealID, error) {
	requestValue := reflect.Indirect(reflect.ValueOf(request))
	taskID := reflect.Indirect(requestValue.FieldByName("Id"))
	if !taskID.IsValid() {
		return "", errNoTaskFieldFound
	}

	if taskID.Type().Kind() != reflect.String {
		return "", errInvalidTaskField
	}

	taskInfo, err := t.hub.getTask(taskID.String())
	if err != nil {
		return "", err
	}

	return DealID(taskInfo.GetDealId()), nil
}
