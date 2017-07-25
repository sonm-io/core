package hub

import (
	"crypto/ecdsa"
	//"fmt"
	"github.com/sonm-io/fusrodah/fusrodah"
	//"github.com/sonm-io/fusrodah/util"
	"github.com/ethereum/go-ethereum/crypto"
	//"github.com/ethereum/go-ethereum/whisper/whisperv2"
	"github.com/ethereum/go-ethereum/whisper/whisperv2"
)

/**
/--------HUB--------/
HUB FUNCTION SECTION
/--------------------/
*/

const Enode = "enode://83985f67557c65877db4098e8640bc8c2c3e903c22c0943c2a895f587ecb5dfc455b2b6855d3ce9466538b8e60be478a15e1ef0f364ddadfcd1ef6754678b292@172.16.1.128:30348"
const DEFAULT_HUB_PORT = ":30322"

type Server struct {
	PrivateKey  *ecdsa.PrivateKey
	Frd         fusrodah.Fusrodah
	KnowingHubs []HubsType
	confFile    string

	HubIp string
}

func NewServer(prv *ecdsa.PrivateKey, hubIp string) *Server {
	if prv == nil {
		//TODO: cover error
		prv, _ = crypto.GenerateKey()
	}

	frd := fusrodah.Fusrodah{
		Prv:   prv,
		Enode: Enode,
		Port:  DEFAULT_HUB_PORT,
	}

	srv := Server{
		PrivateKey: prv,
		HubIp:      hubIp,
		Frd:        frd,
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
	srv.Frd.AddHandling(nil, nil, func(msg *whisperv2.Message) {
		srv.Frd.Send(srv.HubIp, nil, true,  "minerDiscover")
	}, "hubDiscover")
}


func (srv *Server) GetPubKeyString() string {
	pkString := string(crypto.FromECDSAPub(&srv.PrivateKey.PublicKey))
	return pkString
}
