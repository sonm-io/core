package worker

import (
	"context"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kr/pty"
	"github.com/sonm-io/core/insonmnia/npp"
	xssh "github.com/sonm-io/core/insonmnia/ssh"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	"google.golang.org/grpc/credentials"

	sshd "github.com/gliderlabs/ssh"
)

const (
	sshStatusOK          sshStatus = 0
	sshStatusServerError           = 255
)

type sshStatus int

type SSHConfig struct {
	Endpoint  string             `yaml:"endpoint" default:":0"`
	NPP       npp.Config         `yaml:"npp"`
	Identity  sonm.IdentityLevel `yaml:"identity" default:"identified"`
	Blacklist []common.Address   `yaml:"blacklist"`
}

type PublicKey struct {
	ssh.PublicKey
}

func (m *PublicKey) UnmarshalText(data []byte) error {
	if len(data) > 0 {
		pkey, _, _, _, err := ssh.ParseAuthorizedKey(data)
		if err != nil {
			return err
		}
		m.PublicKey = pkey
	}

	return nil
}

func (m PublicKey) MarshalText() ([]byte, error) {
	if m.PublicKey != nil {
		return ssh.MarshalAuthorizedKey(m), nil
	}

	return []byte{}, nil
}

type SSH interface {
	Run(ctx context.Context) error
	Close() error
}

type nilSSH struct{}

func (nilSSH) Run(ctx context.Context) error { return nil }
func (nilSSH) Close() error                  { return nil }

// OverseerView is a bridge between keeping "Worker" as a parameter and
// slightly more decomposed architecture.
type OverseerView interface {
	ContainerInfo(id string) (*ContainerInfo, bool)
	// ConsumerIdentityLevel returns the consumer identity level by the given
	// task identifier.
	ConsumerIdentityLevel(ctx context.Context, id string) (sonm.IdentityLevel, error)
	ExecIdentity() sonm.IdentityLevel
	Exec(ctx context.Context, id string, cmd []string, env []string, isTty bool, wCh <-chan sshd.Window) (types.HijackedResponse, error)
}

func hasHexPrefix(v string) bool {
	return len(v) >= 2 && v[0] == '0' && (v[1] == 'x' || v[1] == 'X')
}

func IsWorkerSSHIdentity(v string) bool {
	if hasHexPrefix(v) {
		v = v[2:]
	}

	if len(v) < 2*common.AddressLength {
		return false
	}

	return common.IsHexAddress(v[:2*common.AddressLength])
}

type sshAuthorizationOptions struct {
	Expiration time.Time
}

func newSSHAuthorizationOptions() *sshAuthorizationOptions {
	return &sshAuthorizationOptions{
		Expiration: time.Unix(math.MaxInt32, 0),
	}
}

type SSHAuthorizationOption func(options *sshAuthorizationOptions)

func WithExpiration(duration time.Duration) SSHAuthorizationOption {
	return func(options *sshAuthorizationOptions) {
		options.Expiration = time.Now().Add(duration)
	}
}

type SSHAuthorization struct {
	mu sync.RWMutex
	// Keys from the blacklist are forbidden for any SSH command.
	//
	// If the same key is appeared both in the whitelist and blacklist the
	// precedence will be for the blacklist.
	deniedKeys map[common.Address]time.Time
	// Keys from the whitelist are always (or temporary) allowed to execute
	// any SSH command.
	allowedKeys map[common.Address]time.Time
}

func NewSSHAuthorization() *SSHAuthorization {
	return &SSHAuthorization{
		deniedKeys:  map[common.Address]time.Time{},
		allowedKeys: map[common.Address]time.Time{},
	}
}

// Allow adds the given key to the whitelist.
func (m *SSHAuthorization) Allow(key common.Address, options ...SSHAuthorizationOption) {
	opts := newSSHAuthorizationOptions()
	for _, o := range options {
		o(opts)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.allowedKeys[key] = opts.Expiration
}

// Deny adds the given key to the blacklist.
func (m *SSHAuthorization) Deny(key common.Address, options ...SSHAuthorizationOption) {
	opts := newSSHAuthorizationOptions()
	for _, o := range options {
		o(opts)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.deniedKeys[key] = opts.Expiration
}

// IsAllowed returns true if the given key passes the authorization.
func (m *SSHAuthorization) IsAllowed(key common.Address) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if expirationTime, ok := m.deniedKeys[key]; ok {
		if time.Now().Before(expirationTime) {
			return false
		}
	}

	if expirationTime, ok := m.allowedKeys[key]; ok {
		return time.Now().Before(expirationTime)
	}

	return false
}

func (m *SSHAuthorization) expirationTimeFromDuration(duration time.Duration) time.Time {
	expirationTime := time.Unix(math.MaxInt32, 0)
	if duration != 0 {
		expirationTime = time.Now().Add(duration)
	}

	return expirationTime
}

type connHandler struct {
	sshAuthorization *SSHAuthorization
	overseer         OverseerView
	log              *zap.SugaredLogger
}

func newConnHandler(sshAuthorization *SSHAuthorization, overseer OverseerView, log *zap.SugaredLogger) *connHandler {
	return &connHandler{
		sshAuthorization: sshAuthorization,
		overseer:         overseer,
		log:              log,
	}
}

func (m *connHandler) Verify(ctx sshd.Context, key sshd.PublicKey) bool {
	user := ctx.User()
	if IsWorkerSSHIdentity(user) {
		// We do not return error otherwise, delaying the check to "process"
		// method to have human-readable errors.
		return true
	}

	if err := m.verifyContainerLogin(user, key); err != nil {
		m.log.Warnw("verification failed", zap.Error(err))
		return false
	}

	return true
}

func (m *connHandler) verifyContainerLogin(taskID string, key sshd.PublicKey) error {
	m.log.Debugf("public key %s verification from user %s", ssh.FingerprintSHA256(key), taskID)

	containerInfo, ok := m.overseer.ContainerInfo(taskID)
	if !ok {
		return fmt.Errorf("container `%s` not found", taskID)
	}

	if !sshd.KeysEqual(containerInfo.PublicKey, key) {
		return fmt.Errorf("provided public key `%s` is not equal with the specified key `%s`", ssh.FingerprintSHA256(key), ssh.FingerprintSHA256(containerInfo.PublicKey))
	}

	return nil
}

func (m *connHandler) onSession(session sshd.Session) {
	status, err := m.process(session)
	if err != nil {
		session.Write([]byte(capitalize(err.Error()) + "\n"))
		m.log.Warnw("failed to process ssh session", zap.Error(err))
	}

	session.Exit(int(status))

	m.log.Infof("finished processing ssh session with %d status", int(status))
}

func (m *connHandler) process(session sshd.Session) (sshStatus, error) {
	m.log.Debugf("processing %v", session.RemoteAddr())

	if IsWorkerSSHIdentity(session.User()) {
		sshIdentity, err := xssh.ParseSSHIdentity(session.User())
		if err != nil {
			return sshStatusServerError, fmt.Errorf("failed to extract SSH identity: %v", err)
		}

		if err := sshIdentity.Verify(); err != nil {
			return sshStatusServerError, fmt.Errorf("failed to verify SSH identity: %v", err)
		}

		if m.sshAuthorization.IsAllowed(sshIdentity.Addr) {
			m.log.Infof("authorized using %s", sshIdentity.Addr.Hex())
			return m.processLogin(session)
		}
		return sshStatusServerError, fmt.Errorf("access denied")
	}

	identity, err := m.overseer.ConsumerIdentityLevel(session.Context(), session.User())
	if err != nil {
		return sshStatusServerError, fmt.Errorf("failed to extract identity level for task `%s`: %v", session.User(), err)
	}

	if identity < m.overseer.ExecIdentity() {
		return sshStatusServerError, fmt.Errorf("identity level `%s` does not allow to exec ssh: must be `%s` or higher", identity.String(), m.overseer.ExecIdentity())
	}

	containerInfo, ok := m.overseer.ContainerInfo(session.User())
	if !ok {
		return sshStatusServerError, fmt.Errorf("failed to find container for task `%s`", session.User())
	}

	cmd := session.Command()
	if len(cmd) == 0 {
		cmd = []string{"login", "-f", "root"}
	}

	_, wCh, isTty := session.Pty()

	stream, err := m.overseer.Exec(session.Context(), containerInfo.ID, cmd, session.Environ(), isTty, wCh)
	if err != nil {
		return sshStatusServerError, err
	}

	defer stream.Close()
	outputErr := make(chan error)

	go func() {
		var err error
		if isTty {
			_, err = io.Copy(session, stream.Reader)
		} else {
			_, err = stdcopy.StdCopy(session, session.Stderr(), stream.Reader)
		}
		outputErr <- err
	}()

	go func() {
		defer stream.CloseWrite()
		io.Copy(stream.Conn, session)
	}()

	if err := <-outputErr; err != nil {
		m.log.Warnw("I/O error during SSH session", zap.Error(err))
		return sshStatusServerError, nil
	}

	return sshStatusOK, nil
}

func (m *connHandler) processLogin(session sshd.Session) (sshStatus, error) {
	command := session.Command()
	if len(command) == 0 {
		sh := os.Getenv("SHELL")
		if len(sh) == 0 {
			sh = "/bin/sh"
		}

		command = []string{sh}
	}

	m.log.Infof("executing `%s` over SSH", strings.Join(command, " "))

	cmd := exec.Command(strings.Join(command, " "))
	cmd.Env = session.Environ()
	ptyReq, winCh, isPty := session.Pty()

	if isPty {
		cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
		fh, err := pty.Start(cmd)
		if err != nil {
			return sshStatusServerError, err
		}

		go func() {
			for win := range winCh {
				setWindowSize(fh, win.Width, win.Height)
			}
		}()

		go func() {
			io.Copy(fh, session)
		}()

		io.Copy(session, fh)
	} else {
		return sshStatusServerError, fmt.Errorf("only shell command currently supported")
	}

	return sshStatusOK, nil
}

type sshServer struct {
	cfg         SSHConfig
	credentials credentials.TransportCredentials
	server      *sshd.Server
	log         *zap.SugaredLogger
}

func NewSSHServer(cfg SSHConfig, signer ssh.Signer, credentials credentials.TransportCredentials, sshAuthorization *SSHAuthorization, overseer OverseerView, log *zap.SugaredLogger) (*sshServer, error) {
	for _, addr := range cfg.Blacklist {
		sshAuthorization.Deny(addr)
	}

	connHandler := newConnHandler(sshAuthorization, overseer, log)

	server := &sshd.Server{
		Handler:          connHandler.onSession,
		HostSigners:      []sshd.Signer{signer},
		PublicKeyHandler: connHandler.Verify,
	}

	m := &sshServer{
		cfg:         cfg,
		credentials: credentials,
		server:      server,
		log:         log,
	}

	return m, nil
}

func (m *sshServer) Run(ctx context.Context) error {
	m.log.Info("running ssh server")
	defer m.log.Info("stopped ssh server")

	listener, err := m.newListener(ctx)
	if err != nil {
		return err
	}
	defer listener.Close()

	return m.server.Serve(listener)
}

func (m *sshServer) newListener(ctx context.Context) (net.Listener, error) {
	nppOptions := []npp.Option{
		npp.WithProtocol("ssh"),
		npp.WithRendezvous(m.cfg.NPP.Rendezvous, m.credentials),
		// TODO: Relay.
		npp.WithLogger(m.log.Desugar()),
	}

	return npp.NewListener(ctx, m.cfg.Endpoint, nppOptions...)
}

func (m *sshServer) Close() error {
	m.log.Info("closing ssh server")

	return m.server.Close()
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}

	return strings.ToUpper(s[:1]) + s[1:]
}

func setWindowSize(fh *os.File, w, h int) {
	syscall.Syscall(
		syscall.SYS_IOCTL,
		fh.Fd(),
		uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})),
	)
}
