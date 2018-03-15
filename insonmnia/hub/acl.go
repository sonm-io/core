package hub

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

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

// workerACLStorage describes an ACL storage for workers.
//
// A worker connection can be accepted only and the only if its credentials
// provided with the certificate contains in this storage.
type workerACLStorage struct {
	storage *set.SetNonTS
}

func newWorkerACLStorage() *workerACLStorage {
	return &workerACLStorage{
		storage: set.NewNonTS(),
	}
}

func (s *workerACLStorage) MarshalJSON() ([]byte, error) {
	if s == nil {
		return json.Marshal(nil)
	}

	set := make([]string, 0)
	s.storage.Each(func(item interface{}) bool {
		set = append(set, item.(string))
		return true
	})

	return json.Marshal(set)
}

func (s *workerACLStorage) UnmarshalJSON(data []byte) error {
	unmarshaled := make([]string, 0)
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		return err
	}
	s.storage = set.NewNonTS()

	for _, val := range unmarshaled {
		s.storage.Add(val)
	}

	return nil
}

// Insert inserts the given worker credentials to the storage.
func (s *workerACLStorage) Insert(credentials string) {
	s.storage.Add(credentials)
}

// Remove removes the given worker credentials from the storage.
// Returns true if it was actually removed.
func (s *workerACLStorage) Remove(credentials string) bool {
	exists := s.storage.Has(credentials)
	if exists {
		s.storage.Remove(credentials)
	}

	return exists
}

// Has checks whether the given worker credentials contains in the
// storage.
func (s *workerACLStorage) Has(credentials string) bool {
	return s.storage.Has(credentials)
}

// Each applies the specified function to each credentials in the storage.
// Traversal will continue until all items in the Set have been visited,
// or if the closure returns false.
func (s *workerACLStorage) Each(fn func(string) bool) {
	s.storage.Each(func(credentials interface{}) bool {
		return fn(credentials.(string))
	})
}

// DealExtractor allows to extract deal id that is used for authorization.
type DealExtractor func(ctx context.Context, request interface{}) (DealID, error)

type dealAuthorization struct {
	ctx       context.Context
	state     *state
	extractor DealExtractor
}

func newDealAuthorization(ctx context.Context, hubState *state, extractor DealExtractor) auth.Authorization {
	return &dealAuthorization{
		ctx:       ctx,
		state:     hubState,
		extractor: extractor,
	}
}

func (d *dealAuthorization) Authorize(ctx context.Context, request interface{}) error {
	dealID, err := d.extractor(ctx, request)
	if err != nil {
		return err
	}

	wallet, err := auth.ExtractWalletFromContext(ctx)
	if err != nil {
		return err
	}

	peerWallet := wallet.Hex()
	meta, err := d.state.GetDealMeta(dealID)

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

// NewFieldDealExtractor constructs a deal id extractor that requires the
// specified request to have "sonm.Deal" field.
// Extraction is performed using reflection.
func newFieldDealExtractor() DealExtractor {
	return func(ctx context.Context, request interface{}) (DealID, error) {
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
}

// NewContextDealExtractor constructs a deal id extractor that requires the
// specified context to have "deal" metadata.
func newContextDealExtractor() DealExtractor {
	return func(ctx context.Context, request interface{}) (DealID, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return "", errNoPeerInfo
		}

		dealMD := md["deal"]
		if len(dealMD) == 0 {
			return "", errNoDealProvided
		}

		return DealID(dealMD[0]), nil
	}
}

// NewFromTaskDealExtractor constructs a deal id extractor that requires the
// specified request to have "Id" field, which is the task id.
// This task id is used to extract current deal id from the Hub.
func newFromTaskDealExtractor(hubState *state) DealExtractor {
	return newFromNamedTaskDealExtractor(hubState, "Id")
}

func newFromNamedTaskDealExtractor(hubState *state, name string) DealExtractor {
	return func(ctx context.Context, request interface{}) (DealID, error) {
		requestValue := reflect.Indirect(reflect.ValueOf(request))
		taskID := reflect.Indirect(requestValue.FieldByName(name))
		if !taskID.IsValid() {
			return "", errNoTaskFieldFound
		}

		if taskID.Type().Kind() != reflect.String {
			return "", errInvalidTaskField
		}

		taskInfo, ok := hubState.GetTaskByID(taskID.String())
		if !ok {
			return "", status.Errorf(codes.NotFound, "task %s not found", taskID.String())
		}

		return DealID(taskInfo.GetDealId()), nil
	}
}

func newRequestDealExtractor(fn func(request interface{}) (DealID, error)) DealExtractor {
	return newCustomDealExtractor(func(ctx context.Context, request interface{}) (DealID, error) {
		return fn(request)
	})
}

func newCustomDealExtractor(fn func(ctx context.Context, request interface{}) (DealID, error)) DealExtractor {
	return func(ctx context.Context, request interface{}) (DealID, error) {
		return fn(ctx, request)
	}
}

// OrderExtractor allows to extract order id that is used for authorization.
type OrderExtractor func(request interface{}) (OrderID, error)

type orderAuthorization struct {
	state     *state
	extractor OrderExtractor
}

func newOrderAuthorization(hubState *state, extractor OrderExtractor) auth.Authorization {
	return &orderAuthorization{
		state:     hubState,
		extractor: extractor,
	}
}

func (a *orderAuthorization) Authorize(ctx context.Context, request interface{}) error {
	wallet, err := auth.ExtractWalletFromContext(ctx)
	if err != nil {
		return err
	}

	orderID, err := a.extractor(request)
	if err != nil {
		return err
	}

	if err := a.state.PollCommitOrder(orderID, *wallet); err != nil {
		return status.Errorf(codes.Unauthenticated, "%s", err)
	}

	return nil
}

type multiAuth struct {
	authorizers []auth.Authorization
}

func (a *multiAuth) Authorize(ctx context.Context, request interface{}) error {
	for _, au := range a.authorizers {
		if err := au.Authorize(ctx, request); err == nil {
			return nil
		}
	}

	return errors.New("all of required auth methods is failed")
}

func newMultiAuth(a ...auth.Authorization) auth.Authorization {
	return &multiAuth{authorizers: a}
}
