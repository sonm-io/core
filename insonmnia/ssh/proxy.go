package ssh

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	sshd "github.com/gliderlabs/ssh"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/npp"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/sync/errgroup"
)

const (
	proto            = "ssh"
	sshAgentSockName = "SSH_AUTH_SOCK"
)

type connMeta struct {
	Addr     *auth.Addr
	Identity string
}

type NilSSHProxyServer struct{}

func (m *NilSSHProxyServer) Serve(ctx context.Context) error {
	return nil
}

type SSHProxyServer struct {
	cfg        ProxyServerConfig
	privateKey *ecdsa.PrivateKey
	market     blockchain.MarketAPI
	options    []npp.Option
	log        *zap.SugaredLogger
}

func convertHostSigners(v []ssh.Signer) []sshd.Signer {
	var converted []sshd.Signer
	for id := range v {
		converted = append(converted, v[id])
	}

	return converted
}

// NewSSHProxyServer constructs a new SSH proxy server that will serve SSH
// connections in remote containers by smart-forwarding traffic via itself.
//
// The server requires SSH agent running on the host system with appropriate
// keys loaded in it. While running it will NOT modify the data within the
// agent.
//
// Example of external usage: "ssh <DealID>.<TaskID>@<host> -p <port>".
func NewSSHProxyServer(cfg ProxyServerConfig, privateKey *ecdsa.PrivateKey, credentials *xgrpc.TransportCredentials, market blockchain.MarketAPI, log *zap.SugaredLogger) (*SSHProxyServer, error) {
	options := []npp.Option{
		npp.WithProtocol(proto),
		npp.WithRendezvous(cfg.NPP.Rendezvous, credentials),
		// TODO: Activate relay, but for now disable for rendezvous testing.
		npp.WithLogger(log.Desugar()),
	}

	m := &SSHProxyServer{
		cfg:        cfg,
		privateKey: privateKey,
		market:     market,
		options:    options,
		log:        log,
	}

	return m, nil
}

// Serve starts serving the SSH proxy server until the specified context is
// canceled or a critical error occurs.
func (m *SSHProxyServer) Serve(ctx context.Context) error {
	m.log.Infof("exposing SSH server on %s", m.cfg.Addr)
	defer m.log.Infof("stopped SSH server on %s", m.cfg.Addr)

	agentSock, err := net.Dial("unix", os.Getenv(sshAgentSockName))
	if err != nil {
		return fmt.Errorf("failed to open ssh agent socket: %v", err)
	}
	defer agentSock.Close()

	hostSigners, err := agent.NewClient(agentSock).Signers()
	if err != nil {
		return fmt.Errorf("failed to extract signers from ssh agent: %v", err)
	}
	if len(hostSigners) == 0 {
		return fmt.Errorf("failed to extract signers from ssh agent: no identities known to the agent")
	}

	nppDialer, err := npp.NewDialer(m.options...)
	if err != nil {
		return err
	}
	defer nppDialer.Close()

	connHandler := &connHandler{
		market:      m.market,
		nppDialer:   nppDialer,
		privateKey:  m.privateKey,
		hostSigners: hostSigners,
		log:         m.log,
	}

	server := &sshd.Server{
		Addr:        m.cfg.Addr,
		Handler:     connHandler.onHandle,
		HostSigners: convertHostSigners(hostSigners),
		PublicKeyHandler: func(ctx sshd.Context, key sshd.PublicKey) bool {
			for _, signer := range hostSigners {
				if sshd.KeysEqual(signer.PublicKey(), key) {
					return true
				}
			}

			return false
		},
	}
	defer server.Close()

	listener, err := net.Listen("tcp", m.cfg.Addr)
	if err != nil {
		return err
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		return server.Serve(listener)
	})

	<-ctx.Done()

	// Close this listener explicitly, otherwise race is possible, when the
	// context is cancelled before SSH server started.
	listener.Close()

	return wg.Wait()
}

type connHandler struct {
	market      blockchain.MarketAPI
	nppDialer   *npp.Dialer
	privateKey  *ecdsa.PrivateKey
	hostSigners []ssh.Signer
	log         *zap.SugaredLogger
}

func (m *connHandler) onHandle(session sshd.Session) {
	defer session.Close()

	if err := m.handle(session); err != nil {
		session.Write(formatErr(err))
		session.Exit(1)
		return
	}

	session.Exit(0)
}

func (m *connHandler) handle(session sshd.Session) error {
	m.log.Infof("accepted SSH connection from %s", session.RemoteAddr())

	pty, windows, isTty := session.Pty()
	m.log.Debugw("handling SSH connection",
		zap.Bool("tty", isTty),
		zap.String("user", session.User()),
		zap.String("terminal", pty.Term),
		zap.String("publicKey", safeFingerprintSHA256(session.PublicKey())),
	)

	meta, err := m.extractMeta(session)
	if err != nil {
		return err
	}

	m.log.Debugf("resolved remote: %s", meta.Addr.String())

	conn, err := m.nppDialer.Dial(*meta.Addr)
	if err != nil {
		return err
	}

	m.log.Debugf("connected to remote endpoint %s", conn.RemoteAddr())

	cfg := &ssh.ClientConfig{
		User: meta.Identity,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(m.hostSigners...),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	clientConn, channels, requests, err := ssh.NewClientConn(conn, meta.Addr.String(), cfg)
	if err != nil {
		return err
	}

	client := ssh.NewClient(clientConn, channels, requests)
	defer client.Close()

	remoteSession, err := client.NewSession()
	if err != nil {
		return err
	}
	defer remoteSession.Close()

	if isTty {
		if err := remoteSession.RequestPty(pty.Term, pty.Window.Height, pty.Window.Width, ssh.TerminalModes{}); err != nil {
			return err
		}
	}

	for _, env := range session.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid environment variable: %s", env)
		}

		remoteSession.Setenv(parts[0], parts[1])
	}

	stdin, err := remoteSession.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := remoteSession.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := remoteSession.StderrPipe()
	if err != nil {
		return err
	}
	if err := remoteSession.Start(strings.Join(session.Command(), " ")); err != nil {
		return err
	}

	m.log.Infof("opened SSH tunnel between %s -> %s, version %s", session.LocalAddr(), clientConn.RemoteAddr(), clientConn.ClientVersion())

	forwardFunc := func(direction string, rd io.Reader, wr io.Writer) error {
		m.log.Infof("forwarding %s", direction)

		written, err := io.Copy(wr, rd)
		if err != nil {
			m.log.With(zap.Error(err)).Warnf("failed to forward %s", direction)
			return err
		}

		m.log.Infof("finished forwarding %s: %d bytes written", direction, written)

		return nil
	}

	wg, ctx := errgroup.WithContext(session.Context())
	go func() error {
		return forwardFunc("-> stdin", session, stdin)
	}()
	// TODO: stdout/stderr intermixing is possible. How to get with it?
	wg.Go(func() error {
		return forwardFunc("<- stdout", stdout, session)
	})
	wg.Go(func() error {
		return forwardFunc("<- stderr", stderr, session)
	})
	go func() error {
		for window := range windows {
			m.log.Debugf("detected window change: %dx%d", window.Height, window.Width)
			if err := remoteSession.WindowChange(window.Height, window.Width); err != nil {
				return err
			}
		}

		return nil
	}()
	go func() error {
		// When we're closing session first.
		<-ctx.Done()
		remoteSession.Close()
		return nil
	}()
	wg.Go(func() error {
		// When remote session is finished.
		err := remoteSession.Wait()
		return err
	})

	return wg.Wait()
}

func (m *connHandler) extractMeta(session sshd.Session) (*connMeta, error) {
	if common.IsHexAddress(session.User()) {
		identity, err := NewSSHIdentity(m.privateKey)
		if err != nil {
			return nil, err
		}

		return &connMeta{
			Addr:     auth.NewETHAddr(common.HexToAddress(session.User())),
			Identity: identity.String(),
		}, nil
	}

	user, err := parseUserIdentity(session.User())
	if err != nil {
		return nil, err
	}

	m.log.Debugw("resolving worker remote using passed user identity", zap.Any("user", user))
	addr, err := m.resolve(session.Context(), user.DealID)
	if err != nil {
		return nil, err
	}

	return &connMeta{
		Addr:     addr,
		Identity: user.TaskID,
	}, nil
}

func (m *connHandler) resolve(ctx context.Context, dealID *big.Int) (*auth.Addr, error) {
	deal, err := m.market.GetDealInfo(ctx, dealID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve `%s` deal into ETH address: %v", dealID.String(), err)
	}

	if deal.Status == sonm.DealStatus_DEAL_CLOSED {
		return nil, fmt.Errorf("failed to resolve `%s` deal into ETH address: deal is closed", dealID.String())
	}

	return auth.NewAddr(deal.GetSupplierID().Unwrap().Hex())
}

type userIdentity struct {
	DealID *big.Int
	TaskID string
}

func parseUserIdentity(user string) (*userIdentity, error) {
	parts := strings.Split(user, ".")

	if len(parts) != 2 {
		return nil, fmt.Errorf("user identity must be in format <DealID>.<TaskID>, but received `%s`", user)
	}

	dealID, ok := new(big.Int).SetString(parts[0], 10)
	if !ok {
		return nil, fmt.Errorf("deal ID must be a number, but received `%s`", parts[0])
	}

	return &userIdentity{
		DealID: dealID,
		TaskID: parts[1],
	}, nil
}

func formatErr(err error) []byte {
	return []byte(fmt.Sprintf("Failed to ssh: %s.\n", err.Error()))
}

func safeFingerprintSHA256(publicKey ssh.PublicKey) string {
	if publicKey == nil {
		return ""
	}

	return ssh.FingerprintSHA256(publicKey)
}
