package node

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/insonmnia/matcher"
	"github.com/sonm-io/core/insonmnia/npp"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
)

type workerClientCreator func(ctx context.Context, addr *auth.Addr) (*workerClient, io.Closer, error)

type workerClient struct {
	sonm.WorkerClient
	sonm.WorkerManagementClient
	sonm.InspectClient
}

// remoteOptions describe options related to remove gRPC services
type remoteOptions struct {
	cfg *Config
	key *ecdsa.PrivateKey

	eth           blockchain.API
	dwh           sonm.DWHClient
	nppDialer     *npp.Dialer
	workerCreator workerClientCreator

	benchList    benchmarks.BenchList
	orderMatcher matcher.Matcher

	log *zap.SugaredLogger
}

func (re *remoteOptions) getWorkerClientForDeal(ctx context.Context, id string) (*workerClient, io.Closer, error) {
	bigID, err := util.ParseBigInt(id)
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse deal id %s to BigInt: %s", id, err)
	}

	dealInfo, err := re.eth.Market().GetDealInfo(ctx, bigID)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get deal info for deal %s from blockchain: %s", id, err)
	}
	if dealInfo.Status == sonm.DealStatus_DEAL_CLOSED {
		return nil, nil, fmt.Errorf("deal %s is closed", id)
	}

	client, closer, err := re.getWorkerClientByEthAddr(ctx, dealInfo.GetSupplierID().Unwrap().Hex())
	if err != nil {
		return nil, nil, fmt.Errorf("could not get worker client for deal %s by eth address %s: %s",
			id, dealInfo.GetSupplierID().Unwrap().Hex(), err)
	}
	return client, closer, nil
}

func (re *remoteOptions) getWorkerClientByEthAddr(ctx context.Context, eth string) (*workerClient, io.Closer, error) {
	return re.workerCreator(ctx, auth.NewETHAddr(common.HexToAddress(eth)))
}

// isWorkerAvailable building worker client by eth address, then call .Status method
func (re *remoteOptions) isWorkerAvailable(ctx context.Context, addr common.Address) bool {
	worker, closer, err := re.getWorkerClientByEthAddr(ctx, addr.Hex())
	if err != nil {
		return false
	}

	defer closer.Close()

	_, err = worker.Status(ctx, &sonm.Empty{})
	return err == nil
}

func newRemoteOptions(ctx context.Context, cfg *Config, key *ecdsa.PrivateKey, credentials *xgrpc.TransportCredentials, log *zap.SugaredLogger) (*remoteOptions, error) {
	nppDialerOptions := []npp.Option{
		npp.WithRendezvous(cfg.NPP.Rendezvous, credentials),
		npp.WithRelay(cfg.NPP.Relay, key),
		npp.WithLogger(log.Desugar()),
	}
	nppDialer, err := npp.NewDialer(nppDialerOptions...)
	if err != nil {
		return nil, err
	}

	workerFactory := func(ctx context.Context, addr *auth.Addr) (*workerClient, io.Closer, error) {
		if addr == nil {
			return nil, nil, fmt.Errorf("no address specified to dial worker")
		}
		conn, err := nppDialer.DialContext(ctx, *addr)
		if err != nil {
			return nil, nil, err
		}
		ethAddr, err := addr.ETH()
		if err != nil {
			return nil, nil, err
		}

		cc, err := xgrpc.NewClient(ctx, "-", auth.NewWalletAuthenticator(credentials, ethAddr), xgrpc.WithConn(conn))
		if err != nil {
			return nil, nil, err
		}

		m := &workerClient{
			sonm.NewWorkerClient(cc),
			sonm.NewWorkerManagementClient(cc),
			sonm.NewInspectClient(cc),
		}

		return m, cc, nil
	}

	dwhCC, err := xgrpc.NewClient(ctx, cfg.DWH.Endpoint, credentials)
	if err != nil {
		return nil, err
	}

	dwh := sonm.NewDWHClient(dwhCC)

	eth, err := blockchain.NewAPI(ctx, blockchain.WithConfig(cfg.Blockchain), blockchain.WithNiceMarket())
	if err != nil {
		return nil, err
	}

	benchList, err := benchmarks.NewBenchmarksList(ctx, cfg.Benchmarks)
	if err != nil {
		return nil, err
	}

	var orderMatcher matcher.Matcher
	if cfg.Matcher != nil {
		orderMatcher, err = matcher.NewMatcher(&matcher.Config{
			Key:        key,
			DWH:        dwh,
			Eth:        eth,
			PollDelay:  cfg.Matcher.PollDelay,
			QueryLimit: cfg.Matcher.QueryLimit,
			Log:        log,
		})

		if err != nil {
			return nil, err
		}
	} else {
		orderMatcher = matcher.NewDisabledMatcher()
	}

	return &remoteOptions{
		cfg:           cfg,
		key:           key,
		eth:           eth,
		dwh:           dwh,
		nppDialer:     nppDialer,
		workerCreator: workerFactory,
		benchList:     benchList,
		orderMatcher:  orderMatcher,
		log:           log,
	}, nil
}
