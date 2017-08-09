package miner

import (
	"io"
	"net"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gliderlabs/ssh"
	log "github.com/noxiouz/zapctx/ctxlog"
	"go.uber.org/zap"
)

type SSH interface {
	Run(miner *Miner) error
	Close()
}

type sshServer struct {
	miner    *Miner
	laddr    string
	listener net.Listener
	server   *ssh.Server
}

func NewSSH(laddr string) (SSH, error) {
	ret := sshServer{laddr: laddr}
	return &ret, nil
}

func (s *sshServer) Run(miner *Miner) error {
	s.miner = miner
	l, err := net.Listen("tcp", s.laddr)
	if err != nil {
		return err
	}
	s.listener = l
	s.server = &ssh.Server{}
	s.server.Handle(s.onSession)
	s.server.PublicKeyHandler = s.verify
	return s.server.Serve(s.listener)
}

func (s *sshServer) verify(ctx ssh.Context, key ssh.PublicKey) bool {
	return true
}

func (s *sshServer) onSession(session ssh.Session) {
	status := s.process(session)
	session.Exit(status)
	log.G(s.miner.ctx).Info("finished processing ssh session", zap.Int("status", status))
}

func (s *sshServer) process(session ssh.Session) (status int) {
	status = 255
	_, wCh, isTty := session.Pty()

	cmd := session.Command()
	if len(cmd) == 0 {
		cmd = append(cmd, "login", "-f", "root")
	}
	cid, ok := s.miner.getContainerIdByTaskId(session.User())
	if !ok {
		msg := "could not find container by task " + string(session.User()+"\n")
		session.Write([]byte(msg))
		log.G(s.miner.ctx).Warn(msg)
		return
	}
	stream, err := s.miner.ovs.Exec(s.miner.ctx, cid, cmd, session.Environ(), isTty, wCh)
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
		_, err := io.Copy(stream.Conn, session)
		outputErr <- err
	}()

	err = <-outputErr
	if err != nil {
		status = 0
	} else {
		log.G(s.miner.ctx).Warn("io error during ssh session:", zap.Error(err))
	}
	return
}

func (s *sshServer) Close() {
	if s.server != nil {
		log.G(s.miner.ctx).Info("closing ssh server")
		s.server.Close()
	}
}
