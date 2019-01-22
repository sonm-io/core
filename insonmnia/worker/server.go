package worker

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/cnf/structhash"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	sshd "github.com/gliderlabs/ssh"
	"github.com/gogo/protobuf/proto"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/oschwald/geoip2-golang"
	"github.com/pborman/uuid"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/insonmnia/cgroups"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/insonmnia/hardware/disk"
	"github.com/sonm-io/core/insonmnia/inspect"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/matcher"
	"github.com/sonm-io/core/insonmnia/npp"
	"github.com/sonm-io/core/insonmnia/resource"
	"github.com/sonm-io/core/insonmnia/state"
	"github.com/sonm-io/core/insonmnia/structs"
	"github.com/sonm-io/core/insonmnia/worker/gpu"
	"github.com/sonm-io/core/insonmnia/worker/metrics"
	"github.com/sonm-io/core/insonmnia/worker/plugin"
	"github.com/sonm-io/core/insonmnia/worker/salesman"
	"github.com/sonm-io/core/insonmnia/worker/volume"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/debug"
	"github.com/sonm-io/core/util/defergroup"
	"github.com/sonm-io/core/util/multierror"
	"github.com/sonm-io/core/util/netutil"
	"github.com/sonm-io/core/util/xdocker"
	"github.com/sonm-io/core/util/xgrpc"
	"github.com/sonm-io/core/util/xnet"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	workerAPIPrefix       = "/sonm.WorkerManagement/"
	inspectServicePrefix  = "/sonm.Inspect/"
	taskAPIPrefix         = "/sonm.Worker/"
	sshPrivateKeyKey      = "ssh_private_key"
	ethereumPrivateKeyKey = "ethereum_private_key"
	exportKeystorePath    = "/var/lib/sonm/worker_keystore"
)

var (
	workerManagementMethods = []string{
		workerAPIPrefix + "Tasks",
		workerAPIPrefix + "Devices",
		workerAPIPrefix + "FreeDevices",
		workerAPIPrefix + "AskPlans",
		workerAPIPrefix + "CreateAskPlan",
		workerAPIPrefix + "RemoveAskPlan",
		workerAPIPrefix + "PurgeAskPlans",
		workerAPIPrefix + "PurgeAskPlansDetailed",
		workerAPIPrefix + "ScheduleMaintenance",
		workerAPIPrefix + "NextMaintenance",
		workerAPIPrefix + "DebugState",
		workerAPIPrefix + "RemoveBenchmark",
		workerAPIPrefix + "PurgeBenchmarks",
		workerAPIPrefix + "AddCapability",
	}

	inspectMethods = []string{
		inspectServicePrefix + "Config",
		inspectServicePrefix + "OpenFiles",
		inspectServicePrefix + "Network",
		inspectServicePrefix + "HostInfo",
		inspectServicePrefix + "DockerInfo",
		inspectServicePrefix + "DockerNetwork",
		inspectServicePrefix + "DockerVolumes",
		inspectServicePrefix + "WatchLogs",
	}

	leakedInsecureKey = common.HexToAddress("0x8125721c2413d99a33e351e1f6bb4e56b6b633fd")
)

type overseerView struct {
	worker *Worker
}

func (m *overseerView) ContainerInfo(id string) (*ContainerInfo, bool) {
	return m.worker.GetContainerInfo(id)
}

func (m *overseerView) ConsumerIdentityLevel(ctx context.Context, id string) (sonm.IdentityLevel, error) {
	plan, err := m.worker.AskPlanByTaskID(id)
	if err != nil {
		return sonm.IdentityLevel_UNKNOWN, err
	}

	deal, err := m.worker.salesman.Deal(plan.GetDealID())
	if err != nil {
		return sonm.IdentityLevel_UNKNOWN, err
	}

	return m.worker.eth.ProfileRegistry().GetProfileLevel(ctx, deal.GetConsumerID().Unwrap())
}

func (m *overseerView) ExecIdentity() sonm.IdentityLevel {
	return m.worker.cfg.SSH.Identity
}

func (m *overseerView) Exec(ctx context.Context, id string, cmd []string, env []string, isTty bool, wCh <-chan sshd.Window) (types.HijackedResponse, error) {
	return m.worker.ovs.Exec(ctx, id, cmd, env, isTty, wCh)
}

type workerConfigProvider struct {
	cfg *Config
}

func (m *workerConfigProvider) Config() interface{} {
	return m.cfg
}

// Worker holds information about jobs, make orders to Observer and communicates with Worker
type Worker struct {
	ctx     context.Context
	cfg     *Config
	storage *state.Storage

	ovs         Overseer
	ssh         SSH
	key         *ecdsa.PrivateKey
	publicIPs   []string
	benchmarks  benchmarks.BenchList
	eth         blockchain.API
	dwh         sonm.DWHClient
	credentials *xgrpc.TransportCredentials
	certRotator util.HitlessCertRotator
	plugins     *plugin.Repository
	whitelist   Whitelist
	matcher     matcher.Matcher
	version     string

	mu        sync.Mutex
	hardware  *hardware.Hardware
	resources *resource.Scheduler
	salesman  *salesman.Salesman

	eventAuthorization   *auth.AuthRouter
	inspectAuthorization *auth.AnyOfTransportCredentialsAuthorization

	// Maps StartRequest's IDs to containers' IDs
	// TODO: It's doubtful that we should keep this map here instead in the Overseer.
	containers map[string]*ContainerInfo

	controlGroup  cgroups.CGroup
	cGroupManager cgroups.CGroupManager
	listener      *npp.Listener
	externalGrpc  *grpc.Server

	startTime           time.Time
	isMasterConfirmed   bool
	isBenchmarkFinished bool

	// Geolocation info.
	country *geoip2.Country
	// Hardware metrics for various hardware types.
	metrics *metrics.Handler
	// Embedded inspection service.
	*inspect.InspectService
}

func NewWorker(cfg *Config, storage *state.Storage, options ...Option) (*Worker, error) {
	opts := newOptions()
	for _, opt := range options {
		opt(opts)
	}

	m := &Worker{
		cfg:        cfg,
		ctx:        opts.ctx,
		storage:    storage,
		version:    opts.version,
		containers: map[string]*ContainerInfo{},
	}

	dg := defergroup.DeferGroup{}
	dg.Defer(func() {
		m.Close()
	})

	if err := m.init(); err != nil {
		return nil, err
	}

	if err := m.setupGeoIP(); err != nil {
		return nil, err
	}

	if err := m.setupAuthorization(); err != nil {
		return nil, err
	}

	if err := m.setupStatusServer(); err != nil {
		return nil, err
	}

	if err := m.setupMaster(); err != nil {
		return nil, err
	}

	if err := m.setupControlGroup(); err != nil {
		return nil, err
	}

	if err := m.setupHardware(); err != nil {
		return nil, err
	}

	if err := m.runBenchmarks(); err != nil {
		return nil, err
	}
	m.isBenchmarkFinished = true

	if err := m.setupResources(); err != nil {
		return nil, err
	}

	// First we setup salesman here to restore all known ask plans to perform container restoration.
	if err := m.setupSalesman(); err != nil {
		return nil, err
	}

	// At this step, all ask plans should be restored and task resources can be successfully consumed by the scheduler.
	// Note: if some ask plans were removed due to resource lack corresponding containers will be removed in "setupRunningContainers".
	if err := m.setupRunningContainers(); err != nil {
		return nil, err
	}

	if err := m.setupSSH(&overseerView{worker: m}); err != nil {
		return nil, err
	}

	if err := m.setupInspectService(cfg, m.inspectAuthorization, opts.logWatcher); err != nil {
		return nil, err
	}

	dg.CancelExec()

	return m, nil
}

func (m *Worker) init() error {
	if err := m.setupKey(); err != nil {
		return err
	}

	if err := m.setupBlockchainAPI(); err != nil {
		return err
	}

	if err := m.setupPlugins(); err != nil {
		return err
	}

	if err := m.setupMetrics(); err != nil {
		return err
	}

	if err := m.setupCredentials(); err != nil {
		return err
	}

	if err := m.setupDWH(); err != nil {
		return err
	}

	if err := m.setupWhitelist(); err != nil {
		return err
	}

	if err := m.setupMatcher(); err != nil {
		return err
	}

	if err := m.setupBenchmarks(); err != nil {
		return err
	}

	if err := m.setupNetworkOptions(); err != nil {
		return err
	}

	if err := m.setupOverseer(); err != nil {
		return err
	}

	return nil
}

func (m *Worker) setupKey() error {
	if m.key == nil {
		var data []byte
		loaded, err := m.storage.Load(ethereumPrivateKeyKey, &data)
		if err != nil {
			return err
		}

		if !loaded {
			key, err := crypto.GenerateKey()
			if err != nil {
				return err
			}

			if err := m.storage.Save(ethereumPrivateKeyKey, crypto.FromECDSA(key)); err != nil {
				return err
			}

			m.key = key
		} else {
			key, err := crypto.ToECDSA(data)
			if err != nil {
				return err
			}

			m.key = key
		}
	}

	log.S(m.ctx).Infof("worker has the following ETH address: %s", m.ethAddr().Hex())

	if err := m.exportKey(); err != nil {
		return err
	}

	return nil
}

func (m *Worker) exportKey() error {
	if err := os.MkdirAll(exportKeystorePath, 0700); err != nil {
		return err
	}

	ks := keystore.NewKeyStore(exportKeystorePath, keystore.LightScryptN, keystore.LightScryptP)
	if !ks.HasAddress(crypto.PubkeyToAddress(m.key.PublicKey)) {
		_, err := ks.ImportECDSA(m.key, "sonm")
		return err
	}

	return nil
}

func (m *Worker) setupBlockchainAPI() error {
	if m.eth == nil {
		eth, err := blockchain.NewAPI(m.ctx, blockchain.WithConfig(m.cfg.Blockchain), blockchain.WithNiceMarket())
		if err != nil {
			return err
		}

		m.eth = eth
	}

	return nil
}

func (m *Worker) setupPlugins() error {
	plugins, err := plugin.NewRepository(m.ctx, m.cfg.Plugins)
	if err != nil {
		return err
	}

	m.plugins = plugins
	return nil
}

func (m *Worker) setupMetrics() error {
	h, err := metrics.NewHandler(log.G(m.ctx), m.cfg.Plugins.GPUs)
	if err != nil {
		return err
	}

	m.metrics = h
	m.metrics.Run(m.ctx)
	return nil
}

func (m *Worker) setupCredentials() error {
	if m.credentials == nil {
		certRotator, TLSConfig, err := util.NewHitlessCertRotator(m.ctx, m.key)
		if err != nil {
			return err
		}

		m.certRotator = certRotator
		m.credentials = xgrpc.NewTransportCredentials(TLSConfig)
	}

	return nil
}

func (m *Worker) setupDWH() error {
	if m.dwh == nil {
		cc, err := xgrpc.NewClient(m.ctx, m.cfg.DWH.Endpoint, m.credentials)
		if err != nil {
			return err
		}

		m.dwh = sonm.NewDWHClient(cc)
	}

	return nil
}

func (m *Worker) setupWhitelist() error {
	if m.whitelist == nil {
		cfg := m.cfg.Whitelist
		if len(cfg.PrivilegedAddresses) == 0 {
			cfg.PrivilegedAddresses = append(cfg.PrivilegedAddresses, crypto.PubkeyToAddress(m.key.PublicKey).Hex())
			cfg.PrivilegedAddresses = append(cfg.PrivilegedAddresses, m.cfg.Master.Hex())
			if m.cfg.Admin != nil {
				cfg.PrivilegedAddresses = append(cfg.PrivilegedAddresses, m.cfg.Admin.Hex())
			}
		}

		m.whitelist = NewWhitelist(m.ctx, &cfg)
	}

	return nil
}

func (m *Worker) setupMatcher() error {
	if m.matcher == nil {
		if m.cfg.Matcher != nil {
			matcher, err := matcher.NewMatcher(&matcher.Config{
				Key:        m.key,
				DWH:        m.dwh,
				Eth:        m.eth,
				PollDelay:  m.cfg.Matcher.PollDelay,
				QueryLimit: m.cfg.Matcher.QueryLimit,
				Log:        log.S(m.ctx),
			})
			if err != nil {
				return fmt.Errorf("cannot create matcher: %v", err)
			}

			m.matcher = matcher
		} else {
			m.matcher = matcher.NewDisabledMatcher()
		}
	}

	return nil
}

func (m *Worker) setupBenchmarks() error {
	if m.benchmarks == nil {
		benchList, err := benchmarks.NewBenchmarksList(m.ctx, m.cfg.Benchmarks)
		if err != nil {
			return err
		}

		m.benchmarks = benchList
	}

	return nil
}

func (m *Worker) setupNetworkOptions() error {
	// Use public IPs from config (if provided).
	publicIPs := m.cfg.PublicIPs
	if len(publicIPs) > 0 {
		m.publicIPs = netutil.SortedIPs(publicIPs)
		return nil
	}

	// Scan interfaces if there's no config and no NAT.
	rawPublicIPs, err := netutil.GetPublicIPs()
	if err != nil {
		return err
	}

	for _, ip := range rawPublicIPs {
		publicIPs = append(publicIPs, ip.String())
	}
	m.publicIPs = netutil.SortedIPs(publicIPs)

	return nil
}

func encodeRSAPrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	privateKeyData := x509.MarshalPKCS1PrivateKey(privateKey)
	block := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privateKeyData,
	}

	return pem.EncodeToMemory(&block)
}

func (m *Worker) setupSSH(view OverseerView) error {
	if m.cfg.SSH != nil {
		signer, err := m.loadOrGenerateSSHSigner()
		if err != nil {
			return err
		}

		sshAuthorization := NewSSHAuthorization()
		sshAuthorization.Deny(leakedInsecureKey)
		sshAuthorization.Allow(crypto.PubkeyToAddress(m.key.PublicKey))
		sshAuthorization.Allow(m.cfg.Master)

		if m.cfg.Admin != nil {
			sshAuthorization.Allow(*m.cfg.Admin)
		}

		ssh, err := NewSSHServer(*m.cfg.SSH, signer, m.credentials, sshAuthorization, view, log.S(m.ctx))
		if err != nil {
			return err
		}

		m.ssh = ssh
		return nil
	}

	if m.ssh == nil {
		m.ssh = nilSSH{}
	}

	return nil
}

func (m *Worker) setupInspectService(cfg *Config, authWatcher *auth.AnyOfTransportCredentialsAuthorization, loggingWatcher *logging.Watcher) error {
	inspectService, err := inspect.NewInspectService(&workerConfigProvider{cfg: cfg}, authWatcher, loggingWatcher)
	if err != nil {
		return err
	}

	m.InspectService = inspectService

	return nil
}

func (m *Worker) loadOrGenerateSSHSigner() (ssh.Signer, error) {
	var privateKeyData []byte
	ok, err := m.storage.Load(sshPrivateKeyKey, &privateKeyData)
	if err != nil {
		return nil, err
	}

	if ok {
		return ssh.ParsePrivateKey(privateKeyData)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	if err := m.storage.Save(sshPrivateKeyKey, encodeRSAPrivateKeyToPEM(privateKey)); err != nil {
		return nil, err
	}

	return ssh.NewSignerFromSigner(privateKey)
}

func (m *Worker) setupOverseer() error {
	if m.ovs == nil {
		ovs, err := NewOverseer(m.ctx, m.plugins)
		if err != nil {
			return err
		}

		m.ovs = ovs
	}

	return nil
}

// Serve starts handling incoming API gRPC requests
func (m *Worker) Serve() error {
	m.startTime = time.Now()
	if err := m.waitMasterApproved(); err != nil {
		m.Close()
		return err
	}

	if err := m.setupServer(); err != nil {
		m.Close()
		return err
	}

	log.S(m.ctx).Info("updated hardware in blockchain")

	wg, ctx := errgroup.WithContext(m.ctx)
	wg.Go(func() error {
		return m.RunSSH(ctx)
	})
	wg.Go(func() error {
		log.S(m.ctx).Infof("listening for gRPC API connections on %s", m.listener.Addr())
		defer log.S(m.ctx).Infof("finished listening for gRPC API connections on %s", m.listener.Addr())

		return m.externalGrpc.Serve(m.listener)
	})

	<-ctx.Done()
	m.Close()

	return wg.Wait()
}

func (m *Worker) waitMasterApproved() error {
	if m.cfg.Development.DisableMasterApproval {
		log.S(m.ctx).Debug("skip waiting for master approval: disabled")
		return nil
	}

	if m.isMasterConfirmed {
		// master confirmation is detected at startup, we don't want to check it twice.
		log.S(m.ctx).Debug("skip waiting for master approval: already confirmed")
		return nil
	}

	selfAddr := m.ethAddr().Hex()
	log.S(m.ctx).Infof("waiting approval for %s from master %s", selfAddr, m.cfg.Master.Hex())

	expectedMaster := m.cfg.Master.Hex()
	ticker := util.NewImmediateTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return m.ctx.Err()
		case <-ticker.C:
			addr, err := m.eth.Market().GetMaster(m.ctx, m.ethAddr())
			if err != nil {
				log.S(m.ctx).Warnf("failed to get master: %s, retrying...", err)
				continue
			}

			curMaster := addr.Hex()
			if curMaster == selfAddr {
				log.S(m.ctx).Infof("still no approval for %s from %s, continue waiting", m.ethAddr().Hex(), m.cfg.Master.Hex())
				continue
			}

			if curMaster != expectedMaster {
				return fmt.Errorf("received unexpected master %s", curMaster)
			}

			m.isMasterConfirmed = true
			return nil
		}
	}
}

func (m *Worker) ethAddr() common.Address {
	return crypto.PubkeyToAddress(m.key.PublicKey)
}

func (m *Worker) setupMaster() error {
	if m.cfg.Development.DisableMasterApproval {
		return nil
	}

	log.S(m.ctx).Info("checking current master")
	addr, err := m.eth.Market().GetMaster(m.ctx, m.ethAddr())
	if err != nil {
		return err
	}

	if addr.Big().Cmp(m.ethAddr().Big()) == 0 {
		log.S(m.ctx).Infof("master is not confirmed or not set, sending request from %s to %s",
			m.ethAddr().Hex(), m.cfg.Master.Hex())
		err = m.eth.Market().RegisterWorker(m.ctx, m.key, m.cfg.Master)
		if err != nil {
			return err
		}
	}

	// master is confirmed when expected master addr is equal to existing
	// addr recorded into blockchain.
	m.isMasterConfirmed = m.cfg.Master.Big().Cmp(addr.Big()) == 0
	return nil
}

func (m *Worker) setupGeoIP() error {
	var publicGeoIP net.IP
	for _, publicIP := range m.publicIPs {
		if ip := net.ParseIP(publicIP); ip != nil {
			publicGeoIP = ip
			break
		}
	}

	if publicGeoIP == nil {
		log.G(m.ctx).Info("no public IP detected in network interfaces, nor specified in the config, falling back to external resolver")

		ip, err := xnet.NewExternalPublicIPResolver("").PublicIP()
		if err != nil {
			return fmt.Errorf("failed to detect at least one public IP address")
		}

		log.S(m.ctx).Infof("successfully resolved public IP: %s", ip.String())

		publicGeoIP = ip
	}

	geoIPService, err := NewGeoIPService(&GeoIPServiceConfig{})
	if err != nil {
		return err
	}

	country, err := geoIPService.Country(publicGeoIP)
	if err != nil {
		return fmt.Errorf("failed to detect machine's country by geo IP: %v", err)
	}

	m.country = country

	return nil
}

func (m *Worker) setupAuthorization() error {
	inspectAuthorization := auth.NewAnyOfTransportCredentialsAuthorization(m.ctx)
	inspectAuthorization.Add(m.ethAddr(), time.Duration(0))
	inspectAuthorization.Add(m.cfg.Master, time.Duration(0))

	inspectAuthOptions := []auth.Authorization{
		inspectAuthorization,
	}

	managementAuthOptions := []auth.Authorization{
		auth.NewTransportAuthorization(m.ethAddr()),
		auth.NewTransportAuthorization(m.cfg.Master),
	}

	if m.cfg.Admin != nil {
		inspectAuthOptions = append(inspectAuthOptions, auth.NewTransportAuthorization(*m.cfg.Admin))
		managementAuthOptions = append(managementAuthOptions, auth.NewTransportAuthorization(*m.cfg.Admin))
	}

	inspectAuth := newAnyOfAuth(inspectAuthOptions...)
	managementAuth := newAnyOfAuth(managementAuthOptions...)

	// master, admin, and metrics collector service is allowed to obtain metrics
	metricsCollectorAuth := managementAuthOptions
	if m.cfg.MetricsCollector != nil {
		metricsCollectorAuth = append(metricsCollectorAuth, auth.NewTransportAuthorization(*m.cfg.MetricsCollector))
	}

	authorization := auth.NewEventAuthorization(m.ctx,
		auth.WithLog(log.G(m.ctx)),
		// Note: need to refactor auth router to support multiple prefixes for methods.
		auth.Allow(workerManagementMethods...).With(managementAuth),
		// Everyone can get worker's status.
		auth.Allow(workerAPIPrefix+"Status").With(auth.NewNilAuthorization()),
		auth.Allow(taskAPIPrefix+"TaskStatus").With(newAnyOfAuth(
			managementAuth,
			newDealAuthorization(m.ctx, m, newFromTaskDealExtractor(m)),
		)),
		auth.Allow(taskAPIPrefix+"StopTask").With(newDealAuthorization(m.ctx, m, newFromTaskDealExtractor(m))),
		auth.Allow(taskAPIPrefix+"JoinNetwork").With(newDealAuthorization(m.ctx, m, newFromNamedTaskDealExtractor(m, "TaskID"))),
		auth.Allow(taskAPIPrefix+"StartTask").With(newDealAuthorization(m.ctx, m, newRequestDealExtractor(func(request interface{}) (*sonm.BigInt, error) {
			return request.(*sonm.StartTaskRequest).GetDealID(), nil
		}))),
		auth.Allow(taskAPIPrefix+"PurgeTasks").With(newDealAuthorization(m.ctx, m, newRequestDealExtractor(func(request interface{}) (*sonm.BigInt, error) {
			return request.(*sonm.PurgeTasksRequest).GetDealID(), nil
		}))),
		auth.Allow(taskAPIPrefix+"TaskLogs").With(newDealAuthorization(m.ctx, m, newFromTaskDealExtractor(m))),
		auth.Allow(taskAPIPrefix+"PushTask").With(newAllOfAuth(
			newDealAuthorization(m.ctx, m, newContextDealExtractor()),
			newKYCAuthorization(m.ctx, m.cfg.Whitelist.PrivilegedIdentityLevel, m.eth.ProfileRegistry())),
		),
		auth.Allow(taskAPIPrefix+"PullTask").With(newDealAuthorization(m.ctx, m, newRequestDealExtractor(func(request interface{}) (*sonm.BigInt, error) {
			return sonm.NewBigIntFromString(request.(*sonm.PullTaskRequest).GetDealId())
		}))),
		auth.Allow(taskAPIPrefix+"GetDealInfo").With(newDealAuthorization(m.ctx, m, newRequestDealExtractor(func(request interface{}) (*sonm.BigInt, error) {
			return sonm.NewBigIntFromString(request.(*sonm.ID).GetId())
		}))),
		auth.Allow(workerAPIPrefix+"Metrics").With(newAnyOfAuth(metricsCollectorAuth...)),
		auth.Allow(workerAPIPrefix+"Devices").With(newAnyOfAuth(metricsCollectorAuth...)),
		auth.Allow(inspectMethods...).With(inspectAuth),
		auth.WithFallback(auth.NewDenyAuthorization()),
	)

	m.eventAuthorization = authorization
	m.inspectAuthorization = inspectAuthorization
	return nil
}

func (m *Worker) setupControlGroup() error {
	cgName := "sonm-worker-parent"
	cgResources := &specs.LinuxResources{}
	if m.cfg.Resources != nil {
		cgName = m.cfg.Resources.Cgroup
		cgResources = m.cfg.Resources.Resources
	}

	cgroup, cGroupManager, err := cgroups.NewCgroupManager(cgName, cgResources)
	if err != nil {
		return err
	}
	m.controlGroup = cgroup
	m.cGroupManager = cGroupManager
	return nil
}

func (m *Worker) setupHardware() error {
	// TODO: Do all the stuff inside hardware ctor
	hardwareInfo, err := hardware.NewHardware()
	if err != nil {
		return err
	}

	// check if memory is limited into cgroup
	if s, err := m.controlGroup.Stats(); err == nil {
		if s.MemoryLimit != 0 && s.MemoryLimit < hardwareInfo.RAM.Device.Total {
			hardwareInfo.RAM.Device.Available = s.MemoryLimit
		}
	}

	// apply info about GPUs, expose to logs
	m.plugins.ApplyHardwareInfo(hardwareInfo)
	hardwareInfo.SetNetworkIncoming(m.publicIPs)
	//TODO: configurable?
	hardwareInfo.Network.NetFlags.SetOutbound(true)
	m.hardware = hardwareInfo
	return nil
}

func (m *Worker) CancelDealTasks(dealID *sonm.BigInt) error {
	log.S(m.ctx).Debugf("canceling deal's %s tasks", dealID)
	var toDelete []*ContainerInfo

	m.mu.Lock()
	for key, container := range m.containers {
		if container.DealID.Cmp(dealID) == 0 {
			toDelete = append(toDelete, container)
			delete(m.containers, key)
		}
	}
	m.mu.Unlock()

	result := multierror.NewMultiError()
	for _, container := range toDelete {
		if err := m.ovs.OnDealFinish(m.ctx, container.ID); err != nil {
			result = multierror.Append(result, err)
		}
		if err := m.resources.OnDealFinish(container.TaskId); err != nil {
			result = multierror.Append(result, err)
		}
		if _, err := m.storage.Remove(container.ID); err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result.ErrorOrNil()
}

type runningContainerInfo struct {
	Description Description   `json:"description,omitempty"`
	Cinfo       ContainerInfo `json:"cinfo,omitempty"`
	Spec        sonm.TaskSpec `json:"spec,omitempty"`
}

func (m *Worker) saveContainerInfo(id string, info ContainerInfo, d Description, spec sonm.TaskSpec) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.storage.Save(info.ID, runningContainerInfo{
		Description: d,
		Cinfo:       info,
		Spec:        spec,
	})

	m.containers[id] = &info
}

func (m *Worker) GetContainerInfo(id string) (*ContainerInfo, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	info, ok := m.containers[id]
	return info, ok
}

func (m *Worker) Devices(ctx context.Context, request *sonm.Empty) (*sonm.DevicesReply, error) {
	return m.hardware.IntoProto(), nil
}

// Status returns internal worker statistic
func (m *Worker) Status(ctx context.Context, _ *sonm.Empty) (*sonm.StatusReply, error) {
	var uptime uint64
	if m.isBenchmarkFinished {
		uptime = uint64(time.Now().Sub(m.startTime).Seconds())
	}

	rendezvousStatus := "not connected"
	if m.listener != nil {
		nppMetrics := m.listener.Metrics()
		if nppMetrics.RendezvousAddr != nil {
			rendezvousStatus = nppMetrics.RendezvousAddr.String()
		}
	}

	var adminAddr *sonm.EthAddress
	if m.cfg.Admin != nil {
		adminAddr = sonm.NewEthAddress(*m.cfg.Admin)
	}

	reply := &sonm.StatusReply{
		Uptime:              uptime,
		Version:             m.version,
		Platform:            util.GetPlatformName(),
		EthAddr:             m.ethAddr().Hex(),
		TaskCount:           uint32(len(m.CollectTasksStatuses(sonm.TaskStatusReply_RUNNING))),
		DWHStatus:           m.cfg.Endpoint,
		RendezvousStatus:    rendezvousStatus,
		Master:              sonm.NewEthAddress(m.cfg.Master),
		Admin:               adminAddr,
		IsMasterConfirmed:   m.isMasterConfirmed,
		IsBenchmarkFinished: m.isBenchmarkFinished,
		Geo: &sonm.GeoIP{
			Country: &sonm.GeoIPCountry{
				IsoCode: m.country.Country.IsoCode,
			},
		},
	}

	return reply, nil
}

// FreeDevices provides information about unallocated resources
// that can be turned into ask-plans.
// Deprecated: no longer usable
func (m *Worker) FreeDevices(ctx context.Context, request *sonm.Empty) (*sonm.DevicesReply, error) {
	resources, err := m.resources.GetCommitedFree()
	if err != nil {
		return nil, err
	}

	freeHardware, err := m.hardware.LimitTo(resources)
	if err != nil {
		return nil, err
	}

	return freeHardware.IntoProto(), nil
}

func (m *Worker) setStatus(status *sonm.TaskStatusReply, id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.containers[id]
	if !ok {
		m.containers[id] = &ContainerInfo{}
	}

	m.containers[id].status = status.GetStatus()
	if status.Status == sonm.TaskStatusReply_BROKEN || status.Status == sonm.TaskStatusReply_FINISHED {
		m.resources.ReleaseTask(id)
	}
}

func (m *Worker) listenForStatus(statusListener chan sonm.TaskStatusReply_Status, id string) {
	select {
	case newStatus, ok := <-statusListener:
		if !ok {
			return
		}
		m.setStatus(&sonm.TaskStatusReply{Status: newStatus}, id)
	case <-m.ctx.Done():
		return
	}
}

func (m *Worker) PushTask(stream sonm.Worker_PushTaskServer) error {
	if err := m.eventAuthorization.Authorize(stream.Context(), auth.Event(taskAPIPrefix+"PushTask"), nil); err != nil {
		return err
	}

	request, err := structs.NewImagePush(stream)
	if err != nil {
		return err
	}
	log.G(m.ctx).Info("pushing image", zap.Int64("size", request.ImageSize()))

	result, err := m.ovs.Load(stream.Context(), newChunkReader(stream))
	if err != nil {
		return err
	}

	log.G(m.ctx).Info("image loaded, set trailer", zap.String("trailer", result.String()))
	stream.SetTrailer(metadata.Pairs("id", result.String()))
	return nil
}

func (m *Worker) PullTask(request *sonm.PullTaskRequest, stream sonm.Worker_PullTaskServer) error {
	if err := m.eventAuthorization.Authorize(stream.Context(), auth.Event(taskAPIPrefix+"PullTask"), request); err != nil {
		return err
	}

	ctx := log.WithLogger(m.ctx, log.G(m.ctx).With(zap.String("request", "pull task"), zap.String("id", uuid.New())))

	task, err := m.TaskStatus(ctx, &sonm.ID{Id: request.GetTaskId()})
	if err != nil {
		log.G(m.ctx).Warn("could not fetch task history by deal", zap.Error(err))
		return err
	}

	named, err := reference.ParseNormalizedNamed(task.GetImageName())
	if err != nil {
		log.G(m.ctx).Warn("could not parse image to reference", zap.Error(err), zap.String("image", task.GetImageName()))
		return err
	}

	tagged, err := reference.WithTag(named, fmt.Sprintf("%s_%s", request.GetDealId(), request.GetTaskId()))
	if err != nil {
		log.G(m.ctx).Warn("could not tag image", zap.Error(err), zap.String("image", task.GetImageName()))
		return err
	}
	imageID := tagged.String()

	log.G(ctx).Debug("pulling image", zap.String("imageID", imageID))

	info, rd, err := m.ovs.Save(stream.Context(), imageID)
	if err != nil {
		return err
	}
	defer rd.Close()

	stream.SendHeader(metadata.Pairs("size", strconv.FormatInt(info.Size, 10)))

	streaming := true
	buf := make([]byte, 1*1024*1024)
	for streaming {
		n, err := rd.Read(buf)
		if err != nil {
			if err == io.EOF {
				streaming = false
			} else {
				return err
			}
		}
		if err := stream.Send(&sonm.Chunk{Chunk: buf[:n]}); err != nil {
			return err
		}
	}

	return nil
}

func (m *Worker) taskAllowed(ctx context.Context, request *sonm.StartTaskRequest) (bool, xdocker.Reference, error) {
	spec := request.GetSpec()
	ref, err := xdocker.NewReference(spec.GetContainer().GetImage())
	if err != nil {
		return false, ref, fmt.Errorf("failed to parse reference: %s", err)
	}

	deal, err := m.salesman.Deal(request.GetDealID())
	if err != nil {
		return false, ref, err
	}
	level, err := m.eth.ProfileRegistry().GetProfileLevel(ctx, deal.GetConsumerID().Unwrap())
	if err != nil {
		return false, ref, err
	}
	if level < m.cfg.Whitelist.PrivilegedIdentityLevel {
		volumes := spec.GetContainer().GetVolumes()
		mounts := spec.GetContainer().GetMounts()
		if len(volumes) > 0 || len(mounts) > 0 {
			return false, ref, fmt.Errorf("mounting volumes is forbidden due to kyc level")
		}
		return m.whitelist.Allowed(ctx, ref, spec.GetRegistry().Auth())
	}

	return true, ref, nil
}

func (m *Worker) StartTask(ctx context.Context, request *sonm.StartTaskRequest) (*sonm.StartTaskReply, error) {
	allowed, ref, err := m.taskAllowed(ctx, request)
	if err != nil {
		return nil, err
	}

	if !allowed {
		return nil, status.Errorf(codes.PermissionDenied, "specified image is forbidden to run")
	}

	taskID := uuid.New()

	dealID := request.GetDealID()
	ask, err := m.salesman.AskPlanByDeal(dealID)
	if err != nil {
		return nil, err
	}

	cgroup, err := m.salesman.CGroup(ask.ID)
	if err != nil {
		return nil, err
	}

	spec := request.GetSpec()

	publicKey := PublicKey{}
	err = publicKey.UnmarshalText([]byte(spec.GetContainer().GetSshKey()))
	if err != nil {
		return nil, fmt.Errorf("failed to parse SSH public key: %v", err)
	}

	network, err := m.salesman.Network(ask.ID)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid public key provided %v", err)
	}
	if spec.GetResources() == nil {
		spec.Resources = &sonm.AskPlanResources{}
	}
	if spec.GetResources().GetGPU() == nil {
		spec.Resources.GPU = ask.Resources.GPU
	}

	hasher := &sonm.AskPlanHasher{AskPlanResources: ask.GetResources()}
	err = spec.GetResources().GetGPU().Normalize(hasher)
	if err != nil {
		log.G(ctx).Error("could not normalize GPU resources", zap.Error(err))
		m.setStatus(&sonm.TaskStatusReply{Status: sonm.TaskStatusReply_BROKEN}, taskID)
		return nil, status.Errorf(codes.Internal, "could not normalize GPU resources: %s", err)
	}

	//TODO: generate ID
	if err := m.resources.ConsumeTask(ask.ID, taskID, spec.Resources); err != nil {
		return nil, fmt.Errorf("could not start task: %s", err)
	}

	mounts := make([]volume.Mount, 0)
	for _, spec := range spec.Container.Mounts {
		mount, err := volume.NewMount(spec)
		if err != nil {
			m.resources.ReleaseTask(taskID)
			return nil, err
		}
		mounts = append(mounts, mount)
	}

	networks, err := structs.NewNetworkSpecs(spec.Container.Networks)
	if err != nil {
		log.G(ctx).Error("failed to parse networking specification", zap.Error(err))
		m.setStatus(&sonm.TaskStatusReply{Status: sonm.TaskStatusReply_BROKEN}, taskID)
		return nil, status.Errorf(codes.Internal, "failed to parse networking specification: %s", err)
	}
	gpuids, err := m.hardware.GPUIDs(spec.GetResources().GetGPU())
	if err != nil {
		log.G(ctx).Error("failed to fetch GPU IDs ", zap.Error(err))
		m.setStatus(&sonm.TaskStatusReply{Status: sonm.TaskStatusReply_BROKEN}, taskID)
		return nil, status.Errorf(codes.Internal, "failed to fetch GPU IDs: %s", err)
	}

	if len(spec.GetContainer().GetExpose()) > 0 {
		if !ask.GetResources().GetNetwork().GetNetFlags().GetIncoming() {
			m.setStatus(&sonm.TaskStatusReply{Status: sonm.TaskStatusReply_BROKEN}, taskID)
			return nil, fmt.Errorf("incoming network is required due to explicit `expose` settings, but not allowed for `%s` deal", dealID.Unwrap())
		}
	}

	var d = Description{
		Container:      *request.Spec.Container,
		Reference:      ref,
		Auth:           spec.Registry.Auth(),
		CGroupParent:   cgroup.Suffix(),
		Resources:      spec.Resources,
		DealId:         request.GetDealID().Unwrap().String(),
		TaskId:         taskID,
		GPUDevices:     gpuids,
		mounts:         mounts,
		NetworkOptions: network,
		NetworkSpecs:   networks,
	}

	// TODO: Detect whether it's the first time allocation. If so - release resources on error.

	m.setStatus(&sonm.TaskStatusReply{Status: sonm.TaskStatusReply_SPOOLING}, taskID)
	log.G(m.ctx).Info("spooling an image")
	if err := m.ovs.Spool(ctx, d); err != nil {
		log.G(ctx).Error("failed to Spool an image", zap.Error(err))
		m.setStatus(&sonm.TaskStatusReply{Status: sonm.TaskStatusReply_BROKEN}, taskID)
		return nil, status.Errorf(codes.Internal, "failed to Spool %v", err)
	}
	log.G(m.ctx).Info("spooled an image")

	m.setStatus(&sonm.TaskStatusReply{Status: sonm.TaskStatusReply_SPAWNING}, taskID)
	log.G(m.ctx).Info("spawning an image")
	statusListener, containerInfo, err := m.ovs.Start(m.ctx, d)
	if err != nil {
		log.G(ctx).Error("failed to spawn an image", zap.Error(err))
		m.setStatus(&sonm.TaskStatusReply{Status: sonm.TaskStatusReply_BROKEN}, taskID)
		return nil, status.Errorf(codes.Internal, "failed to Spawn %v", err)
	}

	log.G(m.ctx).Info("spawned an image")
	containerInfo.PublicKey = publicKey
	containerInfo.StartAt = time.Now()
	containerInfo.ImageName = ref.String()
	containerInfo.DealID = dealID
	containerInfo.Tag = request.GetSpec().GetTag()
	containerInfo.TaskId = taskID
	containerInfo.AskID = ask.ID

	var reply = sonm.StartTaskReply{
		Id:         taskID,
		PortMap:    make(map[string]*sonm.Endpoints, 0),
		NetworkIDs: containerInfo.NetworkIDs,
	}

	for internalPort, portBindings := range containerInfo.Ports {
		if len(portBindings) < 1 {
			continue
		}

		var socketAddrs []*sonm.SocketAddr
		var pubPortBindings []nat.PortBinding

		for _, portBinding := range portBindings {
			hostPort := portBinding.HostPort
			hostPortInt, err := nat.ParsePort(hostPort)
			if err != nil {
				m.resources.ReleaseTask(taskID)
				return nil, err
			}

			for _, publicIP := range m.publicIPs {
				socketAddrs = append(socketAddrs, &sonm.SocketAddr{
					Addr: publicIP,
					Port: uint32(hostPortInt),
				})

				pubPortBindings = append(pubPortBindings, nat.PortBinding{HostIP: publicIP, HostPort: hostPort})
			}
		}

		containerInfo.Ports[internalPort] = pubPortBindings

		reply.PortMap[string(internalPort)] = &sonm.Endpoints{Endpoints: socketAddrs}
	}

	m.saveContainerInfo(taskID, containerInfo, d, *spec)

	go m.listenForStatus(statusListener, taskID)

	deal, err := m.salesman.Deal(dealID)
	if err != nil || deal.Status != sonm.DealStatus_DEAL_ACCEPTED {
		log.G(m.ctx).Warn("deal was closed before task was spawned")
		if err := m.CancelDealTasks(dealID); err != nil {
			log.S(m.ctx).Errorf("failed to drop tasks of closed deals: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to start task: corresponding deal was closed")
	}

	return &reply, nil
}

// StopTask request forces to kill container
func (m *Worker) StopTask(ctx context.Context, request *sonm.ID) (*sonm.Empty, error) {
	m.mu.Lock()
	containerInfo, ok := m.containers[request.Id]
	m.mu.Unlock()

	if !ok {
		return nil, status.Errorf(codes.NotFound, "no job with id %s", request.Id)
	}

	if err := m.stopTask(ctx, containerInfo.ID); err != nil {
		log.G(ctx).Error("failed to Stop container", zap.Error(err))
		return nil, err
	}

	return &sonm.Empty{}, nil
}

func (m *Worker) stopTask(ctx context.Context, id string) error {
	if err := m.ovs.Stop(ctx, id); err != nil {
		m.setStatus(&sonm.TaskStatusReply{Status: sonm.TaskStatusReply_BROKEN}, id)
		return status.Errorf(codes.Internal, "failed to stop container %v", err)
	}

	m.setStatus(&sonm.TaskStatusReply{Status: sonm.TaskStatusReply_FINISHED}, id)
	return nil
}

func (m *Worker) PurgeTasks(ctx context.Context, request *sonm.PurgeTasksRequest) (*sonm.ErrorByStringID, error) {
	var toDelete []*ContainerInfo

	m.mu.Lock()
	for _, task := range m.containers {
		if task.DealID.Cmp(request.GetDealID()) == 0 && task.status == sonm.TaskStatusReply_RUNNING {
			toDelete = append(toDelete, task)
		}
	}
	m.mu.Unlock()

	errs := &sonm.ErrorByStringID{Response: []*sonm.ErrorByStringID_Item{}}
	for _, task := range toDelete {
		item := &sonm.ErrorByStringID_Item{ID: task.TaskId}
		if err := m.stopTask(ctx, task.ID); err != nil {
			item.Error = err.Error()
		}
		errs.Response = append(errs.Response, item)
	}

	return errs, nil
}

func (m *Worker) Tasks(ctx context.Context, request *sonm.Empty) (*sonm.TaskListReply, error) {
	return &sonm.TaskListReply{Info: m.CollectTasksStatuses()}, nil
}

func (m *Worker) CollectTasksStatuses(statuses ...sonm.TaskStatusReply_Status) map[string]*sonm.TaskStatusReply {
	result := map[string]*sonm.TaskStatusReply{}
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, info := range m.containers {
		if len(statuses) > 0 {
			for _, s := range statuses {
				if s == info.status {
					result[id] = info.IntoProto(m.ctx)
					break
				}
			}
		} else {
			result[id] = info.IntoProto(m.ctx)
		}
	}
	return result
}

// TaskLogs returns logs from container
func (m *Worker) TaskLogs(request *sonm.TaskLogsRequest, server sonm.Worker_TaskLogsServer) error {
	if err := m.eventAuthorization.Authorize(server.Context(), auth.Event(taskAPIPrefix+"TaskLogs"), request); err != nil {
		return err
	}
	containerInfo, ok := m.GetContainerInfo(request.Id)
	if !ok {
		return status.Errorf(codes.NotFound, "no job with id %s", request.Id)
	}
	opts := types.ContainerLogsOptions{
		ShowStdout: request.Type == sonm.TaskLogsRequest_STDOUT || request.Type == sonm.TaskLogsRequest_BOTH,
		ShowStderr: request.Type == sonm.TaskLogsRequest_STDERR || request.Type == sonm.TaskLogsRequest_BOTH,
		Since:      request.Since,
		Timestamps: request.AddTimestamps,
		Follow:     request.Follow,
		Tail:       request.Tail,
		Details:    request.Details,
	}
	reader, err := m.ovs.Logs(server.Context(), containerInfo.ID, opts)
	if err != nil {
		return err
	}
	defer reader.Close()
	buffer := make([]byte, 100*1024)
	for {
		readCnt, err := reader.Read(buffer)
		if readCnt != 0 {
			server.Send(&sonm.TaskLogsChunk{Data: buffer[:readCnt]})
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

//TODO: proper request
func (m *Worker) JoinNetwork(ctx context.Context, request *sonm.WorkerJoinNetworkRequest) (*sonm.NetworkSpec, error) {
	spec, err := m.plugins.JoinNetwork(request.NetworkID)
	if err != nil {
		return nil, err
	}
	return &sonm.NetworkSpec{
		Type:    spec.Type,
		Options: spec.Options,
		Subnet:  spec.Subnet,
		Addr:    spec.Addr,
	}, nil
}

func (m *Worker) TaskStatus(ctx context.Context, req *sonm.ID) (*sonm.TaskStatusReply, error) {
	log.G(m.ctx).Info("starting TaskDetails status server")

	info, ok := m.GetContainerInfo(req.GetId())
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no task with id %s", req.GetId())
	}

	var metric ContainerMetrics
	var resources *sonm.AskPlanResources
	// If a container has been stoped, ovs.Info has no metrics for such container
	if info.status == sonm.TaskStatusReply_RUNNING {
		metrics, err := m.ovs.Info(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "cannot get container metrics: %s", err.Error())
		}

		metric, ok = metrics[info.ID]
		if !ok {
			return nil, status.Errorf(codes.NotFound, "cannot get metrics for container %s", req.GetId())
		}

		resources, err = m.resources.ResourceByTask(req.GetId())
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "cannot get resources for container %s", req.GetId())
		}
	}

	reply := info.IntoProto(m.ctx)
	reply.Usage = metric.Marshal()
	reply.AllocatedResources = resources

	return reply, nil
}

func (m *Worker) RunSSH(ctx context.Context) error {
	return m.ssh.Run(ctx)
}

// RunBenchmarks perform benchmarking of Worker's resources.
func (m *Worker) runBenchmarks() error {
	if m.cfg.Development.DisableBenchmarking {
		log.S(m.ctx).Warn("benchmarking is disabled due to development mode activated")
		return nil
	}

	requiredBenchmarks := m.benchmarks.ByID()
	for _, bench := range requiredBenchmarks {
		err := m.runBenchmark(bench)
		if err != nil {
			log.S(m.ctx).Errorf("failed to process benchmark %s(%d)", bench.GetCode(), bench.GetID())
			return err
		}
		log.S(m.ctx).Debugf("processed benchmark %s(%d)", bench.GetCode(), bench.GetID())
	}
	m.hardware.SetDevicesFromBenches()

	return nil
}

func (m *Worker) setupResources() error {
	m.resources = resource.NewScheduler(m.ctx, m.hardware)
	return nil
}

func (m *Worker) setupSalesman() error {
	s, err := salesman.NewSalesman(
		m.ctx,
		salesman.WithLogger(log.S(m.ctx).With("source", "salesman")),
		salesman.WithStorage(m.storage),
		salesman.WithResources(m.resources),
		salesman.WithHardware(m.hardware),
		salesman.WithEth(m.eth),
		salesman.WithCGroupManager(m.cGroupManager),
		salesman.WithMatcher(m.matcher),
		salesman.WithEthkey(m.key),
		salesman.WithConfig(&m.cfg.Salesman),
		salesman.WithDealDestroyer(m),
	)
	if err != nil {
		return err
	}
	m.salesman = s

	m.salesman.Run(m.ctx)
	return nil
}

func (m *Worker) setupRunningContainers() error {
	dockerClient, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	// Overseer maintains it's own instance of docker.Client
	defer dockerClient.Close()

	containers, err := dockerClient.ContainerList(m.ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return err
	}

	var closedDeals = map[string]*sonm.Deal{}
	for _, container := range containers {
		var info runningContainerInfo

		if _, ok := container.Labels[overseerTag]; ok {
			loaded, err := m.storage.Load(container.ID, &info)

			if err != nil {
				log.S(m.ctx).Warnf("failed to load running container info %s", err)
				continue
			}
			if !loaded {
				log.S(m.ctx).Warnf("loaded an empty container info")
				continue
			}

			contJson, err := dockerClient.ContainerInspect(m.ctx, container.ID)
			if err != nil {
				log.S(m.ctx).Error("failed to inspect container", zap.String("id", container.ID), zap.Error(err))
				return err
			}

			bigDealID, ok := big.NewInt(0).SetString(info.Description.DealId, 10)
			if !ok {
				return fmt.Errorf("failed to parse container's DealID (%s)", info.Description.DealId)
			}

			deal, err := m.eth.Market().GetDealInfo(m.ctx, bigDealID)
			if err != nil {
				return fmt.Errorf("failed to get deal %v status: %v", info.Description.DealId, err)
			}

			if deal.Status == sonm.DealStatus_DEAL_CLOSED {
				log.G(m.ctx).Info("found task assigned to closed deal, going to cancel it",
					zap.String("deal_id", info.Description.DealId), zap.String("task_id", info.Cinfo.TaskId))
				closedDeals[deal.Id.Unwrap().String()] = deal
			}

			// TODO: Match our proto status constants with docker's statuses
			switch contJson.State.Status {
			case "created", "paused", "restarting", "removing":
				info.Cinfo.status = sonm.TaskStatusReply_UNKNOWN
			case "running":
				info.Cinfo.status = sonm.TaskStatusReply_RUNNING
			case "exited":
				info.Cinfo.status = sonm.TaskStatusReply_FINISHED
			case "dead":
				info.Cinfo.status = sonm.TaskStatusReply_BROKEN
			}

			m.containers[info.Cinfo.TaskId] = &info.Cinfo
			mounts := make([]volume.Mount, 0)

			for _, spec := range info.Spec.Container.Mounts {
				mount, err := volume.NewMount(spec)
				if err != nil {
					return err
				}
				mounts = append(mounts, mount)
			}

			info.Description.mounts = mounts

			m.ovs.Attach(m.ctx, container.ID, info.Description)
			m.resources.ConsumeTask(info.Cinfo.AskID, info.Cinfo.TaskId, info.Spec.Resources)
		}
	}

	for _, deal := range closedDeals {
		if err := m.CancelDealTasks(deal.GetId()); err != nil {
			return fmt.Errorf("failed to cancel tasks for deal %s: %v", deal.Id.Unwrap().String(), err)
		}
	}

	return nil
}

func (m *Worker) setupServer() error {
	if m.externalGrpc != nil {
		log.G(m.ctx).Info("stopping previously running gRPC server")
		m.externalGrpc.GracefulStop()
	}

	logger := log.GetLogger(m.ctx)
	m.externalGrpc = m.createServer(logger, m.eventAuthorization)

	sonm.RegisterWorkerServer(m.externalGrpc, m)
	sonm.RegisterWorkerManagementServer(m.externalGrpc, m)
	sonm.RegisterInspectServer(m.externalGrpc, m)
	grpc_prometheus.Register(m.externalGrpc)

	if m.cfg.Debug != nil {
		go debug.ServePProf(m.ctx, *m.cfg.Debug, log.G(m.ctx))
	}

	nppListener, err := npp.NewListener(m.ctx, m.cfg.Endpoint,
		npp.WithNPPBacklog(m.cfg.NPP.Backlog),
		npp.WithNPPBackoff(m.cfg.NPP.MinBackoffInterval, m.cfg.NPP.MaxBackoffInterval),
		npp.WithRendezvous(m.cfg.NPP.Rendezvous, xgrpc.NewTransportCredentials(m.credentials.TLSConfig)),
		npp.WithRelay(m.cfg.NPP.Relay, m.key),
		npp.WithLogger(log.G(m.ctx)),
	)
	if err != nil {
		log.G(m.ctx).Error("failed to create NPP listener", zap.String("address", m.cfg.Endpoint), zap.Error(err))
		return err
	}

	m.listener = nppListener
	return nil
}

func (m *Worker) setupStatusServer() error {
	logger := log.GetLogger(m.ctx)
	m.externalGrpc = m.createServer(logger, newStatusAuthorization(m.ctx))
	sonm.RegisterWorkerManagementServer(m.externalGrpc, m)

	lis, err := net.Listen("tcp", m.cfg.Endpoint)
	if err != nil {
		logger.Error("status server: failed to listen", zap.String("address", m.cfg.Endpoint), zap.Error(err))
		m.Close()
		return err
	}

	go func() {
		logger.Debug("status server: starting", zap.String("address", m.cfg.Endpoint))
		defer logger.Debug("status server: stopped")
		m.externalGrpc.Serve(lis)
	}()

	return nil
}

func (m *Worker) createServer(logger *zap.Logger, authRouter *auth.AuthRouter) *grpc.Server {
	return xgrpc.NewServer(logger,
		xgrpc.Credentials(m.credentials),
		xgrpc.DefaultTraceInterceptor(),
		xgrpc.RequestLogInterceptor([]string{"PushTask", "PullTask"}),
		xgrpc.AuthorizationInterceptor(authRouter),
		xgrpc.VerifyInterceptor(),
		xgrpc.RateLimitInterceptor(m.ctx, 100.0, map[string]float64{
			"/sonm.WorkerManagement/Status": 20.0,
		}),
	)
}

type BenchmarkHasher interface {
	// HardwareHash returns hash of the hardware, empty string means that we need to rebenchmark everytime
	HardwareHash() string
}

type DeviceKeyer interface {
	StorageKey() string
}

func benchKey(bench *sonm.Benchmark, device interface{}) string {
	return deviceKey(device) + "/benchmarks/" + fmt.Sprintf("%x", structhash.Md5(bench, 1))
}

func deviceKey(device interface{}) string {
	if dev, ok := device.(DeviceKeyer); ok {
		return "hardware/" + dev.StorageKey()
	} else {
		return "hardware/" + reflect.TypeOf(device).Elem().Name()
	}
}

func (m *Worker) getCachedValue(bench *sonm.Benchmark, device interface{}) (uint64, error) {
	var hash string
	if dev, ok := device.(BenchmarkHasher); ok {
		hash = dev.HardwareHash()
	} else {
		hash = fmt.Sprintf("%x", structhash.Md5(device, 1))
	}
	if hash == "" {
		return 0, fmt.Errorf("hashing is disabled for device")
	}

	var storedHash string
	loaded, err := m.storage.Load(deviceKey(device), &storedHash)
	if err != nil {
		return 0, err
	}
	if loaded && hash == storedHash {
		var storedValue uint64
		loaded, err := m.storage.Load(benchKey(bench, device), &storedValue)
		if err != nil {
			return 0, err
		}
		if !loaded {
			return 0, errors.New("benchmark value not found")
		}
		return storedValue, nil
	}
	if err := m.storage.Save(deviceKey(device), hash); err != nil {
		return 0, fmt.Errorf("failed to save hardware hash: %s", err)
	}
	return 0, fmt.Errorf("hardware hashes do not match, current %s, stored %s", hash, storedHash)
}

func (m *Worker) dropCachedValue(benchID uint64) error {
	benches := m.benchmarks.ByID()
	if benchID >= uint64(len(benches)) {
		return fmt.Errorf("benchmark with id %d not found", benchID)
	}
	drop := func(bench *sonm.Benchmark, device interface{}) error {
		_, err := m.storage.Remove(benchKey(bench, device))
		return err
	}
	bench := benches[benchID]
	switch bench.GetType() {
	case sonm.DeviceType_DEV_CPU:
		return drop(bench, m.hardware.CPU.Device)
	case sonm.DeviceType_DEV_GPU:
		multi := multierror.NewMultiError()
		for _, dev := range m.hardware.GPU {
			if err := drop(bench, dev.Device); err != nil {
				multi = multierror.Append(multi, err)
			}
		}
		return multi.ErrorOrNil()
	case sonm.DeviceType_DEV_RAM:
		return drop(bench, m.hardware.RAM.Device)
	case sonm.DeviceType_DEV_STORAGE:
		return drop(bench, m.hardware.Storage.Device)
	case sonm.DeviceType_DEV_NETWORK_IN:
		return drop(bench, m.hardware.Network)
	case sonm.DeviceType_DEV_NETWORK_OUT:
		return drop(bench, m.hardware.Network)
	default:
		return fmt.Errorf("unknown device %d", bench.GetType())
	}
}

func (m *Worker) getBenchValue(bench *sonm.Benchmark, device interface{}) (uint64, error) {
	if bench.GetID() == benchmarks.CPUCores {
		return uint64(m.hardware.CPU.Device.Cores), nil
	}
	if bench.GetID() == benchmarks.RamSize {
		return m.hardware.RAM.Device.Total, nil
	}
	if bench.GetID() == benchmarks.StorageSize {
		info, err := disk.FreeDiskSpace(m.ctx)
		if err != nil {
			return 0, err
		}
		return info.FreeBytes, nil
	}
	if bench.GetID() == benchmarks.GPUCount {
		//GPU count is always 1 for each GPU device.
		return uint64(1), nil
	}
	gpuDevice, isGpu := device.(*sonm.GPUDevice)
	if bench.GetID() == benchmarks.GPUMem {
		if !isGpu {
			return uint64(0), fmt.Errorf("invalid device for GPUMem benchmark")
		}
		return gpuDevice.GetMemory(), nil
	}

	val, err := m.getCachedValue(bench, device)
	if err == nil {
		log.S(m.ctx).Debugf("using cached benchmark value for benchmark %s(%d) - %d", bench.GetCode(), bench.GetID(), val)
		return val, nil
	} else {
		log.S(m.ctx).Infof("failed to get cached benchmark value for benchmark %s(%d): %s", bench.GetCode(), bench.GetID(), err)
	}

	if len(bench.GetImage()) != 0 {
		d, err := getDescriptionForBenchmark(bench)
		if err != nil {
			return uint64(0), fmt.Errorf("could not create description for benchmark: %s", err)
		}
		d.Env[benchmarks.CPUCountBenchParam] = fmt.Sprintf("%d", m.hardware.CPU.Device.Cores)

		if isGpu {
			d.Env[benchmarks.GPUVendorParam] = gpuDevice.VendorType().String()
			d.GPUDevices = []gpu.GPUID{gpu.GPUID(gpuDevice.GetID())}
		}
		res, err := m.execBenchmarkContainer(bench, d)
		if err != nil {

			return uint64(0), err
		}
		if err := m.storage.Save(benchKey(bench, device), res.Result); err != nil {
			log.S(m.ctx).Warnf("failed to save benchmark result in %s", benchKey(bench, device))
		}
		return res.Result, nil
	} else {
		log.S(m.ctx).Warnf("skipping benchmark %s (setting explicitly to 0)", bench.Code)
		return uint64(0), nil
	}
}

func (m *Worker) setBenchmark(bench *sonm.Benchmark, device interface{}, benchMap map[uint64]*sonm.Benchmark) error {
	value, err := m.getBenchValue(bench, device)
	if err != nil {
		return err
	}

	clone := proto.Clone(bench).(*sonm.Benchmark)
	clone.Result = value
	benchMap[bench.GetID()] = clone
	return nil
}

func (m *Worker) runBenchmark(bench *sonm.Benchmark) error {
	log.S(m.ctx).Debugf("processing benchmark %s(%d)", bench.GetCode(), bench.GetID())
	switch bench.GetType() {
	case sonm.DeviceType_DEV_CPU:
		return m.setBenchmark(bench, m.hardware.CPU.Device, m.hardware.CPU.Benchmarks)
	case sonm.DeviceType_DEV_RAM:
		return m.setBenchmark(bench, m.hardware.RAM.Device, m.hardware.RAM.Benchmarks)
	case sonm.DeviceType_DEV_STORAGE:
		return m.setBenchmark(bench, m.hardware.Storage.Device, m.hardware.Storage.Benchmarks)
	case sonm.DeviceType_DEV_NETWORK_IN:
		return m.setBenchmark(bench, m.hardware.Network, m.hardware.Network.BenchmarksIn)
	case sonm.DeviceType_DEV_NETWORK_OUT:
		return m.setBenchmark(bench, m.hardware.Network, m.hardware.Network.BenchmarksOut)
	case sonm.DeviceType_DEV_GPU:
		//TODO: use context to prevent useless benchmarking in case of error
		group := errgroup.Group{}
		for _, dev := range m.hardware.GPU {
			g := dev
			group.Go(func() error {
				return m.setBenchmark(bench, g.Device, g.Benchmarks)
			})
		}
		if err := group.Wait(); err != nil {
			return err
		}
	default:
		log.S(m.ctx).Warnf("invalid benchmark type %d", bench.GetType())
	}
	return nil
}

// execBenchmarkContainerWithResults executes benchmark as docker image,
// returns JSON output with measured values.
func (m *Worker) execBenchmarkContainerWithResults(d Description) (map[string]*benchmarks.ResultJSON, error) {
	logTime := time.Now().Add(-time.Minute)
	err := m.ovs.Spool(m.ctx, d)
	if err != nil {
		return nil, err
	}

	statusChan, statusReply, err := m.ovs.Start(m.ctx, d)
	if err != nil {
		return nil, fmt.Errorf("cannot start container with benchmark: %v", err)
	}
	log.S(m.ctx).Debugf("started benchmark container %s", statusReply.ID)
	defer m.ovs.OnDealFinish(m.ctx, statusReply.ID)

	select {
	case s := <-statusChan:
		if s == sonm.TaskStatusReply_FINISHED || s == sonm.TaskStatusReply_BROKEN {
			log.S(m.ctx).Debugf("benchmark container %s finished", statusReply.ID)
			logOpts := types.ContainerLogsOptions{
				ShowStdout: true,
				//ShowStderr: true,
				Follow: true,
				Since:  strconv.FormatInt(logTime.Unix(), 10),
			}

			reader, err := m.ovs.Logs(m.ctx, statusReply.ID, logOpts)
			if err != nil {
				return nil, fmt.Errorf("cannot create container log reader for %s: %v", statusReply.ID, err)
			}
			log.S(m.ctx).Debugf("requested container %s logs", statusReply.ID)
			defer reader.Close()

			stdoutBuf := bytes.Buffer{}
			stderrBuf := bytes.Buffer{}

			if _, err := stdcopy.StdCopy(&stdoutBuf, &stderrBuf, reader); err != nil {
				return nil, fmt.Errorf("cannot read logs into buffer: %v", err)
			}
			resultsMap, err := parseBenchmarkResult(stdoutBuf.Bytes())
			if err != nil {
				return nil, fmt.Errorf("cannot parse benchmark result: %v", err)
			}

			return resultsMap, nil
		} else {
			return nil, fmt.Errorf("invalid status %d received", s)
		}
	case <-m.ctx.Done():
		return nil, m.ctx.Err()
	}
}

func (m *Worker) execBenchmarkContainer(ben *sonm.Benchmark, des Description) (*benchmarks.ResultJSON, error) {
	log.G(m.ctx).Debug("starting containered benchmark", zap.Any("benchmark", ben))
	res, err := m.execBenchmarkContainerWithResults(des)
	if err != nil {
		return nil, err
	}

	log.G(m.ctx).Debug("received raw benchmark results",
		zap.Uint64("bench_id", ben.GetID()),
		zap.Any("result", res))

	v, ok := res[ben.GetCode()]
	if !ok {
		return nil, fmt.Errorf("no results for benchmark id=%v found into container's output", ben.GetID())
	}

	return v, nil
}

func parseBenchmarkResult(data []byte) (map[string]*benchmarks.ResultJSON, error) {
	v := &benchmarks.ContainerBenchmarkResultsJSON{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		return nil, fmt.Errorf("failed to parse `%s` to json: %s", string(data), err)
	}

	if len(v.Results) == 0 {
		return nil, errors.New("results is empty")
	}

	return v.Results, nil
}

func getDescriptionForBenchmark(b *sonm.Benchmark) (Description, error) {
	ref, err := xdocker.NewReference(b.GetImage())
	if err != nil {
		return Description{}, err
	}
	return Description{
		Reference: ref,
		Container: sonm.Container{Env: map[string]string{
			benchmarks.BenchIDEnvParamName: fmt.Sprintf("%d", b.GetID()),
		}},
	}, nil
}

func (m *Worker) AskPlans(ctx context.Context, _ *sonm.Empty) (*sonm.AskPlansReply, error) {
	return &sonm.AskPlansReply{AskPlans: m.salesman.AskPlans()}, nil
}

func (m *Worker) CreateAskPlan(ctx context.Context, request *sonm.AskPlan) (*sonm.ID, error) {
	if len(request.GetID()) != 0 || !request.GetOrderID().IsZero() || !request.GetDealID().IsZero() {
		return nil, errors.New("creating ask plans with predefined id, order_id or deal_id is not supported")
	}
	if request.GetCreateTime().Unix().UnixNano() != 0 || request.GetLastOrderPlacedTime().Unix().UnixNano() != 0 {
		return nil, errors.New("creating ask plans with predefined timestamps is not supported")
	}
	id, err := m.salesman.CreateAskPlan(request)
	if err != nil {
		return nil, err
	}

	return &sonm.ID{Id: id}, nil
}

func (m *Worker) RemoveAskPlan(ctx context.Context, request *sonm.ID) (*sonm.Empty, error) {
	if err := m.salesman.RemoveAskPlan(ctx, request.GetId()); err != nil {
		return nil, err
	}
	return &sonm.Empty{}, nil
}

func (m *Worker) PurgeAskPlans(ctx context.Context, _ *sonm.Empty) (*sonm.Empty, error) {
	m.salesman.PurgeAskPlans(ctx)
	return &sonm.Empty{}, nil
}

func (m *Worker) PurgeAskPlansDetailed(ctx context.Context, _ *sonm.Empty) (*sonm.ErrorByStringID, error) {
	return m.salesman.PurgeAskPlans(ctx)
}

func (m *Worker) ScheduleMaintenance(ctx context.Context, timestamp *sonm.Timestamp) (*sonm.Empty, error) {
	if err := m.salesman.ScheduleMaintenance(timestamp.Unix()); err != nil {
		return nil, err
	}
	return &sonm.Empty{}, nil
}

func (m *Worker) NextMaintenance(ctx context.Context, _ *sonm.Empty) (*sonm.Timestamp, error) {
	ts := m.salesman.NextMaintenance()
	return &sonm.Timestamp{
		Seconds: ts.Unix(),
	}, nil
}

func (m *Worker) DebugState(ctx context.Context, _ *sonm.Empty) (*sonm.DebugStateReply, error) {
	return &sonm.DebugStateReply{
		SchedulerData: m.resources.DebugDump(),
		SalesmanData:  m.salesman.DebugDump(),
	}, nil
}

func (m *Worker) GetDealInfo(ctx context.Context, id *sonm.ID) (*sonm.DealInfoReply, error) {
	dealID, err := sonm.NewBigIntFromString(id.Id)
	if err != nil {
		return nil, err
	}
	return m.getDealInfo(dealID)
}

func (m *Worker) RemoveBenchmark(ctx context.Context, id *sonm.NumericID) (*sonm.Empty, error) {
	err := m.dropCachedValue(id.Id)
	if err != nil {
		return nil, err
	}
	return &sonm.Empty{}, nil
}

func (m *Worker) PurgeBenchmarks(ctx context.Context, _ *sonm.Empty) (*sonm.Empty, error) {
	multi := multierror.NewMultiError()
	list := m.benchmarks.ByID()
	for id := range list {
		if err := m.dropCachedValue(uint64(id)); err != nil {
			multi = multierror.Append(multi, err)
		}
	}
	return &sonm.Empty{}, multi.ErrorOrNil()
}

func (m *Worker) getDealInfo(dealID *sonm.BigInt) (*sonm.DealInfoReply, error) {
	deal, err := m.salesman.Deal(dealID)
	if err != nil {
		return nil, err
	}

	ask, err := m.salesman.AskPlanByDeal(dealID)
	if err != nil {
		return nil, err
	}
	resources := ask.GetResources()

	running := map[string]*sonm.TaskStatusReply{}
	completed := map[string]*sonm.TaskStatusReply{}

	m.mu.Lock()
	defer m.mu.Unlock()

	for id, c := range m.containers {
		// task is ours
		if c.DealID.Cmp(dealID) == 0 {
			task := c.IntoProto(m.ctx)

			// task is running or preparing to start
			if c.status == sonm.TaskStatusReply_SPOOLING ||
				c.status == sonm.TaskStatusReply_SPAWNING ||
				c.status == sonm.TaskStatusReply_RUNNING {
				running[id] = task
			} else {
				completed[id] = task
			}
		}
	}

	reply := &sonm.DealInfoReply{
		Deal:      deal,
		Running:   running,
		Completed: completed,
		Resources: resources,
	}
	if resources.GetNetwork().GetNetFlags().GetIncoming() {
		reply.PublicIPs = m.publicIPs
	}

	return reply, nil
}

func (m *Worker) AskPlanByTaskID(taskID string) (*sonm.AskPlan, error) {
	planID, err := m.resources.AskPlanIDByTaskID(taskID)
	if err != nil {
		return nil, err
	}
	return m.salesman.AskPlan(planID)
}

func (m *Worker) Metrics(ctx context.Context, req *sonm.WorkerMetricsRequest) (*sonm.WorkerMetricsResponse, error) {
	return m.metrics.Get(), nil
}

func (m *Worker) AddCapability(ctx context.Context, request *sonm.WorkerAddCapabilityRequest) (*sonm.WorkerAddCapabilityResponse, error) {
	switch request.GetScope() {
	case sonm.CapabilityScope_CAPABILITY_NONE:
	case sonm.CapabilityScope_CAPABILITY_SSH:
		return nil, fmt.Errorf("not implemented yet")
	case sonm.CapabilityScope_CAPABILITY_INSPECTION:
		m.inspectAuthorization.Add(request.GetSubject().Unwrap(), time.Duration(request.GetTtl())*time.Second)
	default:
		return nil, fmt.Errorf("unknown scope: %d", request.GetScope())
	}

	return &sonm.WorkerAddCapabilityResponse{}, nil
}

func (m *Worker) RemoveCapability(ctx context.Context, request *sonm.WorkerRemoveCapabilityRequest) (*sonm.WorkerRemoveCapabilityResponse, error) {
	switch request.GetScope() {
	case sonm.CapabilityScope_CAPABILITY_NONE:
	case sonm.CapabilityScope_CAPABILITY_INSPECTION:
		m.inspectAuthorization.Remove(request.GetSubject().Unwrap())
	default:
		return nil, fmt.Errorf("unknown scope: %d", request.GetScope())
	}

	return &sonm.WorkerRemoveCapabilityResponse{}, nil
}

// Close disposes all resources related to the Worker
func (m *Worker) Close() {
	log.G(m.ctx).Info("closing worker")

	if m.ssh != nil {
		m.ssh.Close()
	}
	if m.ovs != nil {
		m.ovs.Close()
	}
	if m.salesman != nil {
		m.salesman.Close()
	}
	if m.plugins != nil {
		m.plugins.Close()
	}
	if m.externalGrpc != nil {
		m.externalGrpc.Stop()
	}
	if m.certRotator != nil {
		m.certRotator.Close()
	}
	m.InspectService.Close()
}

func newStatusAuthorization(ctx context.Context) *auth.AuthRouter {
	return auth.NewEventAuthorization(ctx,
		auth.WithLog(log.G(ctx)),
		auth.Allow(workerAPIPrefix+"Status").With(auth.NewNilAuthorization()),
		auth.WithFallback(auth.NewDenyAuthorization()),
	)
}
