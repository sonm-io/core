package hub

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/whisper/whisperv2"
	"github.com/sonm-io/core/common"
	"github.com/sonm-io/core/fusrodah"
)

const defaultHubPort = ":30322"

type Server struct {
	PrivateKey  *ecdsa.PrivateKey
	Frd         *fusrodah.Fusrodah
	KnowingHubs []HubsType

	HubIp string
}

func NewServer(prv *ecdsa.PrivateKey, hubIp string) *Server {
	if prv == nil {
		var err error
		prv, err = crypto.GenerateKey()
		if err != nil {
			panic(err)
		}
	}

	frd := fusrodah.NewServer(prv, defaultHubPort, common.BootNodeAddr)

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
		srv.Frd.Send(srv.HubIp, true, common.TopicMinerDiscover)
	}, common.TopicHubDiscover)
}

func (srv *Server) GetPubKeyString() string {
	pkString := string(crypto.FromECDSAPub(&srv.PrivateKey.PublicKey))
	return pkString
}
