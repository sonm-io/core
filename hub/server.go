package hub

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/sonm-io/fusrodah/fusrodah"
	"github.com/sonm-io/fusrodah/util"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/whisper/whisperv2"
)

/**
/--------HUB--------/
HUB FUNCTION SECTION
/--------------------/
*/

const Enode = "enode://b0605764bd7c6a816c51325a9cb9d414277d639f420f9dc48b20d12c04c33391b0a99cc8c045d7ba4657de0c04e8bb3b0d4b072ca9779167a75761d7c3c18eb0@10.196.131.151:30348"
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
	fmt.Println("Server start")
	srv.Frd.Start()
}

func (srv *Server) Stop() {
	srv.Frd.Stop()
}

func (srv *Server) Serve() {
	srv.discoveryHandling()
}

func (srv *Server) discoveryHandling() {

	srv.Frd.AddHandling(nil, func(msg *whisperv2.Message) {
		receivedPubKey := crypto.ToECDSAPub(msg.Payload)
		fmt.Println("DISCOVERY RESPONSE #1")
		srv.Frd.Send(srv.GetPubKeyString(), receivedPubKey, "miner", "discover")
	}, "hubDiscover")

	srv.Frd.AddHandling(&srv.PrivateKey.PublicKey, func(msg *whisperv2.Message) {
		receivedPubKey := crypto.ToECDSAPub(msg.Payload)
		fmt.Println("DISCOVERY RESPONSE #2")
		srv.Frd.Send(util.GetLocalIP(), receivedPubKey, "miner", "addr")
	}, "srv", "addr")
}

func (srv *Server) GetPubKeyString() string {
	pkString := string(crypto.FromECDSAPub(&srv.PrivateKey.PublicKey))
	return pkString
}
