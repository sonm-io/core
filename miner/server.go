package miner

import (
	"crypto/ecdsa"
	"github.com/sonm-io/fusrodah/fusrodah"
	"github.com/sonm-io/fusrodah/hub"
	"github.com/ethereum/go-ethereum/crypto"
	//"time"
	"github.com/ethereum/go-ethereum/whisper/whisperv2"
	"time"
)

const Enode = "enode://83985f67557c65877db4098e8640bc8c2c3e903c22c0943c2a895f587ecb5dfc455b2b6855d3ce9466538b8e60be478a15e1ef0f364ddadfcd1ef6754678b292@172.16.1.128:30348"
const DEFAULT_MINER_PORT = ":30343"

/**
/--------MAINER--------/
MAINER FUNCTION SECTION
/--------------------/
*/
type Server struct {
	PrivateKey ecdsa.PrivateKey
	Hubs       []hub.HubsType
	Frd        *fusrodah.Fusrodah
	ConfFile   string
	ip         string
}

func NewServer(prv *ecdsa.PrivateKey) *Server {

	if prv == nil {
		//TODO: cover error
		prv, _ = crypto.GenerateKey()
	}

	frd := fusrodah.Fusrodah{
		Prv:   prv,
		Enode: Enode,
		Port:  DEFAULT_MINER_PORT,
	}

	srv := Server{
		PrivateKey: *prv,
		Frd:        &frd,
	}

	return &srv
}

func (srv *Server) Start() {
	srv.Frd.Start()
}

func (srv *Server) Stop() {
	srv.Frd.Stop()
}

func (srv *Server) Serve() {
	srv.discovery()
}

func (srv *Server) discovery() {
	var filterId int

	done := make(chan struct{})

	filterId = srv.Frd.AddHandling(nil, nil, func(msg *whisperv2.Message) {
		srv.ip = string(msg.Payload)
		srv.Frd.RemoveHandling(filterId)
		close(done)
	}, "minerDiscover")

	select {
	case <-done:
		return
	default:
		srv.Frd.Send(srv.GetPubKeyString(), nil, true, "hubDiscover")
		time.Sleep(time.Millisecond * 1000)
	}
}

func (srv *Server) GetHubIp() string {
	if srv.ip == "0.0.0.0" {
		srv.discovery()
	}
	return srv.ip
}

func (srv *Server) GetPubKeyString() string {
	pkString := string(crypto.FromECDSAPub(&srv.PrivateKey.PublicKey))
	return pkString
}
