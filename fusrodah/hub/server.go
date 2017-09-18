package hub

import (
	"crypto/ecdsa"

	"encoding/json"
	"github.com/ethereum/go-ethereum/whisper/whisperv2"
	"github.com/sonm-io/core/common"
	"github.com/sonm-io/core/fusrodah"
)

const defaultHubPort = ":30343"

type Server struct {
	PrivateKey *ecdsa.PrivateKey
	Frd        *fusrodah.Fusrodah

	workerEndpoint string
	clientEndpoint string
}

func NewServer(prv *ecdsa.PrivateKey, workerEndpt, clientEndpt string) (srv *Server, err error) {
	bootnodes := []string{common.BootNodeAddr, common.SecondBootNodeAddr}

	frd, err := fusrodah.NewServer(prv, defaultHubPort, bootnodes)
	if err != nil {
		return nil, err
	}

	srv = &Server{
		PrivateKey:     prv,
		workerEndpoint: workerEndpt,
		clientEndpoint: clientEndpt,
		Frd:            frd,
	}

	return srv, nil
}

func (srv *Server) Start() (err error) {
	err = srv.Frd.Start()
	if err != nil {
		return err
	}
	return nil
}

func (srv *Server) Stop() (err error) {
	err = srv.Frd.Stop()
	if err != nil {
		return err
	}
	return nil
}

func (srv *Server) Serve() {
	srv.discovery()
}

func (srv *Server) discovery() {
	srv.Frd.AddHandling(nil, nil, func(msg *whisperv2.Message) {
		body := srv.marshalDiscoveryMessage()
		srv.Frd.Send(body, false, common.TopicMinerDiscover)
	}, common.TopicHubDiscover)
}

func (srv *Server) marshalDiscoveryMessage() string {
	s := fusrodah.DiscoveryMessage{
		WorkerEndpoint: srv.workerEndpoint,
		ClientEndpoint: srv.clientEndpoint,
	}
	b, _ := json.Marshal(&s)
	return string(b)
}
