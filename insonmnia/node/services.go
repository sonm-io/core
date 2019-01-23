package node

import (
	"context"

	"github.com/sonm-io/core/optimus"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/rest"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type services struct {
	intercepted    *interceptedAPI
	market         sonm.MarketServer
	deals          sonm.DealManagementServer
	tasks          sonm.TaskManagementServer
	master         sonm.MasterManagementServer
	token          sonm.TokenManagementServer
	blacklist      sonm.BlacklistServer
	profile        sonm.ProfilesServer
	monitoring     sonm.MonitoringServer
	orderPredictor *optimus.PredictorService
}

func newServices(options *remoteOptions) *services {
	return &services{
		intercepted:    newInterceptedAPI(options),
		market:         newMarketAPI(options),
		deals:          newDealsAPI(options),
		tasks:          newTasksAPI(options),
		master:         newMasterManagementAPI(options),
		token:          newTokenManagementAPI(options),
		blacklist:      newBlacklistAPI(options),
		profile:        newProfileAPI(options),
		monitoring:     newMonitoringAPI(options),
		orderPredictor: optimus.NewPredictorService(options.cfg.Predictor, options.eth.Market(), options.benchList, options.dwh, options.log),
	}
}

func (m *services) RegisterGRPC(server *grpc.Server) error {
	if server == nil {
		return nil
	}

	sonm.RegisterWorkerManagementServer(server, m.intercepted)
	sonm.RegisterWorkerServer(server, m.intercepted)
	sonm.RegisterDWHServer(server, m.intercepted)
	sonm.RegisterMarketServer(server, m.market)
	sonm.RegisterDealManagementServer(server, m.deals)
	sonm.RegisterTaskManagementServer(server, m.tasks)
	sonm.RegisterMasterManagementServer(server, m.master)
	sonm.RegisterTokenManagementServer(server, m.token)
	sonm.RegisterBlacklistServer(server, m.blacklist)
	sonm.RegisterProfilesServer(server, m.profile)
	sonm.RegisterMonitoringServer(server, m.monitoring)
	if m.orderPredictor != nil {
		sonm.RegisterOrderPredictorServer(server, m.orderPredictor)
	}

	return nil
}

func (m *services) RegisterREST(server *rest.Server) error {
	if server == nil {
		return nil
	}

	if err := server.RegisterService((*sonm.WorkerManagementServer)(nil), m.intercepted); err != nil {
		return err
	}
	if err := server.RegisterService((*sonm.WorkerServer)(nil), m.intercepted); err != nil {
		return err
	}
	if err := server.RegisterService((*sonm.MarketServer)(nil), m.market); err != nil {
		return err
	}
	if err := server.RegisterService((*sonm.DealManagementServer)(nil), m.deals); err != nil {
		return err
	}
	if err := server.RegisterService((*sonm.TaskManagementServer)(nil), m.tasks); err != nil {
		return err
	}
	if err := server.RegisterService((*sonm.MasterManagementServer)(nil), m.master); err != nil {
		return err
	}
	if err := server.RegisterService((*sonm.TokenManagementServer)(nil), m.token); err != nil {
		return err
	}
	if err := server.RegisterService((*sonm.BlacklistServer)(nil), m.blacklist); err != nil {
		return err
	}
	if err := server.RegisterService((*sonm.ProfilesServer)(nil), m.profile); err != nil {
		return err
	}
	if err := server.RegisterService((*sonm.MonitoringServer)(nil), m.monitoring); err != nil {
		return err
	}
	if err := server.RegisterService((*sonm.DWHServer)(nil), m.intercepted); err != nil {
		return err
	}
	if m.orderPredictor != nil {
		if err := server.RegisterService((*sonm.OrderPredictorServer)(nil), m.orderPredictor); err != nil {
			return err
		}
	}

	return nil
}

func (m *services) Interceptor() grpc.UnaryServerInterceptor {
	return m.intercepted.intercept
}

func (m *services) StreamInterceptor() grpc.StreamServerInterceptor {
	return m.intercepted.streamIntercept
}

func (m *services) Run(ctx context.Context) error {
	wg, ctx := errgroup.WithContext(ctx)

	wg.Go(func() error {
		if m.orderPredictor == nil {
			return nil
		}

		return m.orderPredictor.Serve(ctx)
	})

	<-ctx.Done()
	return wg.Wait()
}
