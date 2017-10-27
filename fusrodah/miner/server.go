package miner

import (
	"crypto/ecdsa"
	"time"

	"encoding/json"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/whisper/whisperv2"
	"github.com/sonm-io/core/common"
	"github.com/sonm-io/core/fusrodah"
	"github.com/sonm-io/core/util"
)

const defaultMinerPort = ":30342"

type HubInfo struct {
	Address   string
	PublicKey *ecdsa.PublicKey
}

type Server struct {
	PrivateKey *ecdsa.PrivateKey
	Frd        *fusrodah.Fusrodah
	Hub        *HubInfo
}

func NewServer(prv *ecdsa.PrivateKey) (srv *Server, err error) {

	if prv == nil {
		prv, err = crypto.GenerateKey()
		if err != nil {
			return nil, err
		}
	}

	bootnodes := []string{common.BootNodeAddr, common.SecondBootNodeAddr}

	frd, err := fusrodah.NewServer(prv, defaultMinerPort, bootnodes)
	if err != nil {
		return nil, err
	}

	srv = &Server{
		PrivateKey: prv,
		Frd:        frd,
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
	var filterID int
	done := make(chan struct{})

	filterID = srv.Frd.AddHandling(nil, nil, func(msg *whisperv2.Message) {
		if hubKey := msg.Recover(); hubKey != nil { // skip unauthenticated messages
			disco := srv.unmarshalDiscoveryMessage(msg.Payload)
			srv.Hub = &HubInfo{
				PublicKey: hubKey,
				Address:   disco.WorkerEndpoint,
			}
			srv.Frd.RemoveHandling(filterID)
			close(done)
		}
	}, common.TopicMinerDiscover)

	t := time.NewTicker(time.Second * 1)
	defer t.Stop()
	select {
	case <-done:
		return
	case <-t.C:
		srv.Frd.Send(util.PubKeyToString(srv.PrivateKey.PublicKey), true, common.TopicHubDiscover)
	}
}

func (*Server) unmarshalDiscoveryMessage(body []byte) fusrodah.DiscoveryMessage {
	msg := fusrodah.DiscoveryMessage{}
	err := json.Unmarshal(body, &msg)
	if err != nil {
		// cannot unmarshal, fallback to previous protocol impl - message body = hubIP
		msg.WorkerEndpoint = string(body)
	}

	return msg
}

func (srv *Server) GetHub() *HubInfo {
	if srv.Hub == nil {
		srv.discovery()
	}
	return srv.Hub
}
