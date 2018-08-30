package worker

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/sonm-io/core/insonmnia/npp"
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
	Endpoint       string             `yaml:"endpoint" required:"true"`
	PrivateKeyPath string             `yaml:"private_key_path" required:"true"`
	NPP            npp.Config         `yaml:"npp"`
	Identity       sonm.IdentityLevel `yaml:"identity" default:"identified"`
}

type PublicKey struct {
	ssh.PublicKey
}

func (m *PublicKey) UnmarshalText(data []byte) error {
	pkey, _, _, _, err := ssh.ParseAuthorizedKey(data)
	if err != nil {
		return err
	}
	m.PublicKey = pkey
	return nil
}

func (m PublicKey) MarshalText() ([]byte, error) {
	return ssh.MarshalAuthorizedKey(m), nil
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
	IdentityLevel(id string) (sonm.IdentityLevel, error)
	ExecIdentity() sonm.IdentityLevel
	Exec(ctx context.Context, id string, cmd []string, env []string, isTty bool, wCh <-chan sshd.Window) (types.HijackedResponse, error)
}

type connHandler struct {
	overseer OverseerView
	log      *zap.SugaredLogger
}

func newConnHandler(overseer OverseerView, log *zap.SugaredLogger) *connHandler {
	return &connHandler{
		overseer: overseer,
		log:      log,
	}
}

func (m *connHandler) Verify(ctx sshd.Context, key sshd.PublicKey) bool {
	if err := m.verify(ctx.User(), key); err != nil {
		m.log.Warnw("verification failed", zap.Error(err))
		return false
	}

	return true
}

func (m *connHandler) verify(taskID string, key sshd.PublicKey) error {
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
	_, wCh, isTty := session.Pty()

	cmd := session.Command()
	if len(cmd) == 0 {
		cmd = []string{"login", "-f", "root"}
	}

	identity, err := m.overseer.IdentityLevel(session.User())
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

	err = <-outputErr
	if err != nil {
		m.log.Warnw("I/O error during SSH session", zap.Error(err))
		return sshStatusServerError, nil
	}

	return sshStatusOK, nil
}

type sshServer struct {
	cfg         SSHConfig
	credentials credentials.TransportCredentials
	server      *sshd.Server
	log         *zap.SugaredLogger
}

func NewSSHServer(cfg SSHConfig, credentials credentials.TransportCredentials, overseer OverseerView, log *zap.SugaredLogger) (*sshServer, error) {
	privateKeyData, err := ioutil.ReadFile(cfg.PrivateKeyPath)
	if err != nil {
		return nil, err
	}

	privateKey, err := ssh.ParsePrivateKey(privateKeyData)
	if err != nil {
		return nil, err
	}

	connHandler := newConnHandler(overseer, log)

	server := &sshd.Server{
		Handler:          connHandler.onSession,
		HostSigners:      []sshd.Signer{privateKey},
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

func parsePublicKey(key string) (ssh.PublicKey, error) {
	var publicKey ssh.PublicKey
	if len(key) != 0 {
		k, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
		if err != nil {
			return nil, err
		}
		publicKey = k
	}

	return publicKey, nil
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}

	return strings.ToUpper(s[:1]) + s[1:]
}
