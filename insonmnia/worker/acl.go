package worker

import (
	"context"
	"net/http"
	"reflect"
	"time"

	"github.com/ethereum/go-ethereum/common"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/structs"
	"github.com/sonm-io/core/util"
	// alias here is required for dumb gomock
	"encoding/json"
	"fmt"
	"io"
	"sync"

	sonm "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/multierror"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

var (
	errNoPeerInfo       = status.Error(codes.Unauthenticated, "no peer info")
	errNoDealProvided   = status.Error(codes.Unauthenticated, "no `deal` metadata provided")
	errNoDealFieldFound = status.Error(codes.Internal, "no `Deal` field found")
	errInvalidDealField = status.Error(codes.Internal, "invalid `Deal` field type")
	errNoTaskFieldFound = status.Errorf(codes.Internal, "no task `ID` field found")
	errInvalidTaskField = status.Error(codes.Internal, "invalid task `ID` field type")
)

const superusersURL = "http://localhost:8080/superusers.json"

// DealExtractor allows to extract deal id that is used for authorization.
type DealExtractor func(ctx context.Context, request interface{}) (structs.DealID, error)

type dealAuthorization struct {
	ctx       context.Context
	extractor DealExtractor
	supplier  DealInfoSupplier
}

type DealInfoSupplier interface {
	GetDealInfo(ctx context.Context, id *sonm.ID) (*sonm.DealInfoReply, error)
}

func newDealAuthorization(ctx context.Context, supplier DealInfoSupplier, extractor DealExtractor) auth.Authorization {
	return &dealAuthorization{
		ctx:       ctx,
		supplier:  supplier,
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
	meta, err := d.supplier.GetDealInfo(ctx, &sonm.ID{Id: dealID.String()})
	if err != nil {
		return err
	}

	allowedWallet := meta.Deal.GetConsumerID().Unwrap().String()

	log.G(ctx).Debug("found allowed wallet for a deal",
		zap.Stringer("deal", dealID),
		zap.String("wallet", peerWallet),
		zap.String("allowedWallet", allowedWallet),
	)

	if allowedWallet != peerWallet {
		return status.Errorf(codes.Unauthenticated, "wallet mismatch: %s", peerWallet)
	}

	return nil
}

type kycFetcher interface {
	GetProfileLevel(ctx context.Context, owner common.Address) (sonm.IdentityLevel, error)
}

type kycAuthorization struct {
	fetcher       kycFetcher
	requiredLevel sonm.IdentityLevel
}

func newKYCAuthorization(ctx context.Context, requiredLevel sonm.IdentityLevel, fetcher kycFetcher) *kycAuthorization {
	return &kycAuthorization{
		fetcher:       fetcher,
		requiredLevel: requiredLevel,
	}
}

func (m *kycAuthorization) Authorize(ctx context.Context, request interface{}) error {
	wallet, err := auth.ExtractWalletFromContext(ctx)
	if err != nil {
		return err
	}

	actualLevel, err := m.fetcher.GetProfileLevel(ctx, *wallet)
	if err != nil {
		return err
	}
	if actualLevel < m.requiredLevel {
		return status.Errorf(codes.Unauthenticated, "action not allowed for identity level %s, required %s identity level", actualLevel, m.requiredLevel)
	}
	return nil
}

// NewContextDealExtractor constructs a deal id extractor that requires the
// specified context to have "deal" metadata.
func newContextDealExtractor() DealExtractor {
	return func(ctx context.Context, request interface{}) (structs.DealID, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return "", errNoPeerInfo
		}

		dealMD := md["deal"]
		if len(dealMD) == 0 {
			return "", errNoDealProvided
		}

		return structs.DealID(dealMD[0]), nil
	}
}

// NewFromTaskDealExtractor constructs a deal id extractor that requires the
// specified request to have "Id" field, which is the task id.
// This task id is used to extract current deal id from the Worker.
func newFromTaskDealExtractor(worker *Worker) DealExtractor {
	return newFromNamedTaskDealExtractor(worker, "Id")
}

// todo: do not accept Worker as param, use some interface that have TaskStatus method.
func newFromNamedTaskDealExtractor(worker *Worker, name string) DealExtractor {
	return func(ctx context.Context, request interface{}) (structs.DealID, error) {
		requestValue := reflect.Indirect(reflect.ValueOf(request))
		taskID := reflect.Indirect(requestValue.FieldByName(name))
		if !taskID.IsValid() {
			return "", errNoTaskFieldFound
		}

		if taskID.Type().Kind() != reflect.String {
			return "", errInvalidTaskField
		}

		_, err := worker.TaskStatus(ctx, &sonm.ID{Id: taskID.String()})
		if err != nil {
			return "", status.Errorf(codes.NotFound, "task %s not found", taskID.String())
		}

		askPlan, err := worker.AskPlanByTaskID(taskID.Interface().(string))
		if err != nil {
			return "", status.Errorf(codes.NotFound, "ask plan for task %s not found: %s", taskID.String(), err)
		}
		if askPlan.GetDealID().IsZero() {
			return "", status.Errorf(codes.NotFound, "deal for ask plan %s, task %s not found, probably it has been ended",
				askPlan.GetID(), taskID.String())
		}
		return structs.DealID(askPlan.GetDealID().Unwrap().String()), nil
	}
}

func newRequestDealExtractor(fn func(request interface{}) (structs.DealID, error)) DealExtractor {
	return newCustomDealExtractor(func(ctx context.Context, request interface{}) (structs.DealID, error) {
		return fn(request)
	})
}

func newCustomDealExtractor(fn func(ctx context.Context, request interface{}) (structs.DealID, error)) DealExtractor {
	return func(ctx context.Context, request interface{}) (structs.DealID, error) {
		return fn(ctx, request)
	}
}

type anyOfAuth struct {
	authorizers []auth.Authorization
}

func (a *anyOfAuth) Authorize(ctx context.Context, request interface{}) error {
	errs := multierror.NewMultiError()
	for _, au := range a.authorizers {
		switch err := au.Authorize(ctx, request); err {
		case nil:
			return nil
		default:
			errs = multierror.AppendUnique(errs, err)
		}
	}

	return status.Error(codes.Unauthenticated, errs.Error())
}

func newAnyOfAuth(a ...auth.Authorization) auth.Authorization {
	return &anyOfAuth{authorizers: a}
}

type allOfAuth struct {
	authorizers []auth.Authorization
}

func (a *allOfAuth) Authorize(ctx context.Context, request interface{}) error {
	for _, au := range a.authorizers {
		if err := au.Authorize(ctx, request); err != nil {
			return err
		}
	}
	return nil
}

func newAllOfAuth(a ...auth.Authorization) auth.Authorization {
	return &allOfAuth{authorizers: a}
}

type superuserAuthorization struct {
	mu         sync.RWMutex
	period     time.Duration
	url        string
	superusers []common.Address
}

func newSuperuserAuthorization(ctx context.Context, cfg SuperusersConfig) *superuserAuthorization {
	out := &superuserAuthorization{
		period: cfg.UpdatePeriod,
		url:    superusersURL,
	}
	go out.updateRoutine(ctx)

	return out
}

func (m *superuserAuthorization) updateRoutine(ctx context.Context) error {
	ticker := util.NewImmediateTicker(m.period)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := m.load(ctx, m.url); err != nil {
				log.G(ctx).Error("could not load superusers", zap.Error(err))
			}
		}
	}
}

func (m *superuserAuthorization) load(ctx context.Context, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	log.G(ctx).Info("fetched whitelist")
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to download whitelist - got %s", resp.Status)
	}

	return m.readList(ctx, resp.Body)
}

func (m *superuserAuthorization) readList(ctx context.Context, jsonReader io.Reader) error {
	decoder := json.NewDecoder(jsonReader)
	rawAddresses := make([]string, 0)
	err := decoder.Decode(&rawAddresses)
	if err != nil {
		return fmt.Errorf("could not decode superusers data: %v", err)
	}

	var superusers = make([]common.Address, len(rawAddresses))
	for idx, address := range rawAddresses {
		superuser, err := util.HexToAddress(address)
		if err != nil {
			return fmt.Errorf("failed to decode %s into common.Address: %v", address, err)
		}
		superusers[idx] = superuser
	}

	m.mu.Lock()
	m.superusers = superusers
	m.mu.Unlock()

	return nil
}

func (m *superuserAuthorization) Authorize(ctx context.Context, request interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	peerInfo, ok := peer.FromContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "no peer info")
	}

	switch authInfo := peerInfo.AuthInfo.(type) {
	case auth.EthAuthInfo:
		for _, superuser := range m.superusers {
			if authInfo.Wallet == superuser {
				return nil
			}
		}
		return status.Errorf(codes.Unauthenticated, "the wallet %s has no access", authInfo.Wallet.Hex())
	default:
		return status.Error(codes.Unauthenticated, "unknown auth info")
	}
}
