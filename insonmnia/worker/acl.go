package worker

import (
	"context"
	"reflect"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/structs"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/multierror"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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

// todo: do not accept Worker as param, use some interface that have DealInfo method.
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
