package worker

import (
	"io"
	"io/ioutil"
	"net"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gliderlabs/ssh"
	log "github.com/noxiouz/zapctx/ctxlog"
	"go.uber.org/zap"
	gossh "golang.org/x/crypto/ssh"
)

type SSH interface {
	Run() error
	Close()
}

type nilSSH struct{}

func (nilSSH) Run() error {
	return nil
}

func (nilSSH) Close() {}

type sshServer struct {
	worker         *Worker
	laddr          string
	privateKeyPath string
	listener       net.Listener
	server         *ssh.Server
}

func NewSSH(worker *Worker, config *SSHConfig) (SSH, error) {
	ret := sshServer{
		laddr:          config.BindEndpoint,
		privateKeyPath: config.PrivateKeyPath,
		worker:         worker,
	}
	return &ret, nil
}

func (s *sshServer) Run() error {
	l, err := net.Listen("tcp", s.laddr)
	if err != nil {
		return err
	}
	s.listener = l
	s.server = &ssh.Server{}
	if len(s.privateKeyPath) != 0 {
		pkeyData, err := ioutil.ReadFile(s.privateKeyPath)
		if err != nil {
			return err
		}
		pkey, err := gossh.ParsePrivateKey(pkeyData)
		if err != nil {
			return err
		}
		s.server.HostSigners = append(s.server.HostSigners, pkey)
	}
	s.server.Handle(s.onSession)
	s.server.PublicKeyHandler = s.verify
	return s.server.Serve(s.listener)
}

func (s *sshServer) verify(ctx ssh.Context, key ssh.PublicKey) bool {
	cinfo, ok := s.worker.GetContainerInfo(ctx.User())
	if !ok {
		return false
	}
	log.G(s.worker.ctx).Info("verifying public key")
	return ssh.KeysEqual(cinfo.PublicKey, key)
}

func (s *sshServer) onSession(session ssh.Session) {
	status := s.process(session)
	session.Exit(status)
	log.G(s.worker.ctx).Info("finished processing ssh session", zap.Int("status", status))
}

func (s *sshServer) process(session ssh.Session) (status int) {
	status = 255
	_, wCh, isTty := session.Pty()

	cmd := session.Command()
	if len(cmd) == 0 {
		cmd = append(cmd, "login", "-f", "root")
	}
	cid, ok := s.worker.getContainerIdByTaskId(session.User())
	if !ok {
		msg := "could not find container by task " + string(session.User()+"\n")
		session.Write([]byte(msg))
		log.G(s.worker.ctx).Warn(msg)
		return
	}
	stream, err := s.worker.ovs.Exec(s.worker.ctx, cid, cmd, session.Environ(), isTty, wCh)
	if err != nil {
		session.Write([]byte(err.Error()))
		return
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
		status = 0
	} else {
		log.G(s.worker.ctx).Warn("io error during ssh session:", zap.Error(err))
	}
	return
}

func (s *sshServer) Close() {
	if s.server != nil {
		log.G(s.worker.ctx).Info("closing ssh server")
		s.server.Close()
	}
}

func parsePublicKey(key string) (ssh.PublicKey, error) {
	var publicKey ssh.PublicKey
	if len(key) != 0 {
		var err error
		k, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
		if err != nil {
			return nil, err
		}
		publicKey = k
	}

	return publicKey, nil
}
