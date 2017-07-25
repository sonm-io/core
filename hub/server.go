package hub

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/whisper/whisperv2"
	"github.com/sonm-io/fusrodah/common"
	"github.com/sonm-io/fusrodah/fusrodah"
)

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
		Enode: common.BootNodeAddr,
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
		srv.Frd.Send(srv.HubIp, nil, true, common.TopicMinerDiscover)
	}, common.TopicHubDiscover)
}

func (srv *Server) GetPubKeyString() string {
	pkString := string(crypto.FromECDSAPub(&srv.PrivateKey.PublicKey))
	return pkString
}
