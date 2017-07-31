package miner

import (
	"crypto/ecdsa"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/whisper/whisperv2"
	"github.com/sonm-io/core/common"
	"github.com/sonm-io/core/fusrodah"
)

const defaultMinerPort = ":30342"

type Server struct {
	PrivateKey *ecdsa.PrivateKey
	Frd        *fusrodah.Fusrodah

	hubIp string
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
		hubIp: "0.0.0.0",
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

func (srv *Server) Stop() (err error){
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
		srv.hubIp = string(msg.Payload)
		srv.Frd.RemoveHandling(filterID)
		close(done)
	}, common.TopicMinerDiscover)

	for {
		select {
		case <-done:
			return
		default:
			srv.Frd.Send(srv.GetPubKeyString(), true, common.TopicHubDiscover)
			time.Sleep(time.Millisecond * 1000)
		}
	}
}

func (srv *Server) GetHubIp() string {
	if srv.hubIp == "0.0.0.0" {
		srv.discovery()
	}
	return srv.hubIp
}

func (srv *Server) GetPubKeyString() string {
	pkString := string(crypto.FromECDSAPub(&srv.PrivateKey.PublicKey))
	return pkString
}
