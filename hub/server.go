package hub

import (
	"crypto/ecdsa"
	//"fmt"
	"github.com/sonm-io/fusrodah/fusrodah"
	//"github.com/sonm-io/fusrodah/util"
	"github.com/ethereum/go-ethereum/crypto"
	//"github.com/ethereum/go-ethereum/whisper/whisperv2"
)

/**
/--------HUB--------/
HUB FUNCTION SECTION
/--------------------/
*/

const Enode = "enode://e97d851aa39884a54320539f5dcab2ec6688e66116459e42b6d57c1d0db68107475875ad0d42230d97ee19a96440f7eba3f7273b8072d10afd4032e321a1f456@172.16.1.128:30348"
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
	//srv.discoveryHandling()
}

//func (srv *Server) discoveryHandling() {
//
//	srv.Frd.AddHandling(nil, func(msg *whisperv2.Message) {
//		receivedPubKey := crypto.ToECDSAPub(msg.Payload)
//		fmt.Println("DISCOVERY RESPONSE #1")
//		srv.Frd.Send(srv.GetPubKeyString(), receivedPubKey, "miner", "discover")
//	}, "hubDiscover")
//
//	srv.Frd.AddHandling(&srv.PrivateKey.PublicKey, func(msg *whisperv2.Message) {
//		receivedPubKey := crypto.ToECDSAPub(msg.Payload)
//		fmt.Println("DISCOVERY RESPONSE #2")
//		srv.Frd.Send(util.GetLocalIP(), receivedPubKey, "miner", "addr")
//	}, "srv", "addr")
//}

func (srv *Server) GetPubKeyString() string {
	pkString := string(crypto.FromECDSAPub(&srv.PrivateKey.PublicKey))
	return pkString
}
