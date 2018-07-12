package node

import (
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/rest"
	"google.golang.org/grpc"
)

type services struct {
	worker    sonm.WorkerManagementServer
	market    sonm.MarketServer
	deals     sonm.DealManagementServer
	tasks     sonm.TaskManagementServer
	master    sonm.MasterManagementServer
	token     sonm.TokenManagementServer
	blacklist sonm.BlacklistServer
	profile   sonm.ProfilesServer
}

func newServices(options *remoteOptions) *services {
	return &services{
		worker:    newWorkerAPI(options),
		market:    newMarketAPI(options),
		deals:     newDealsAPI(options),
		tasks:     newTasksAPI(options),
		master:    newMasterManagementAPI(options),
		token:     newTokenManagementAPI(options),
		blacklist: newBlacklistAPI(options),
		profile:   newProfileAPI(options),
	}
}

func (m *services) RegisterGRPC(server *grpc.Server) error {
	if server == nil {
		return nil
	}

	sonm.RegisterWorkerManagementServer(server, m.worker)
	sonm.RegisterMarketServer(server, m.market)
	sonm.RegisterDealManagementServer(server, m.deals)
	sonm.RegisterTaskManagementServer(server, m.tasks)
	sonm.RegisterMasterManagementServer(server, m.master)
	sonm.RegisterTokenManagementServer(server, m.token)
	sonm.RegisterBlacklistServer(server, m.blacklist)
	sonm.RegisterProfilesServer(server, m.profile)

	return nil
}

func (m *services) RegisterREST(server *rest.Server) error {
	if server == nil {
		return nil
	}

	if err := server.RegisterService((*sonm.WorkerManagementServer)(nil), m.worker); err != nil {
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

	return nil
}

func (m *services) Interceptor() grpc.UnaryServerInterceptor {
	return m.worker.(*workerAPI).intercept
}
