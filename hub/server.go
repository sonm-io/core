package hub

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/sonm-io/fusrodah/fusrodah"
	"github.com/sonm-io/fusrodah/util"
	"github.com/sonm-io/go-ethereum/crypto"
	"github.com/sonm-io/go-ethereum/whisper/whisperv2"
	"io/ioutil"
	"os"
)

/**
/--------HUB--------/
HUB FUNCTION SECTION
/--------------------/
*/

const Enode = "enode://81b8db7b071b46bfc8619268606df7edf48cc55f804f52ce6176bbb369cab22af752ce15c622c958f29dd7617c3d1d647f544f93ce5a11f4319334c418340e3c@172.16.1.111:30348"
const DEFAULT_HUB_PORT = ":30344"

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
	}, "hub", "addr")
}

func (srv *Server) Serve() {
	srv.discoveryHandling()
}

//Deprecated
func (hub *Server) loadKnowingHubs() {
	// NOTE: this for test case any
	hub.KnowingHubs = __getHubList()
}

//Deprecated
func (hub *Server) DiscoveryHandling(frd fusrodah.Fusrodah) {
	//this function load knowing hubs and at the same time
	//and print hubs with topics
	frd.AddHandling(nil, func(msg *whisperv2.Message) {
		msgStr := string(msg.Payload)
		if msgStr != "verifyHub" {
			return
		}

		hub.loadKnowingHubs()
		fmt.Println("Server: discovery response")
		hubListString, err := json.Marshal(hub.KnowingHubs)
		if err != nil {
			fmt.Println(err)
			return
		}
		frd.Send(string(hubListString), nil, "hub", "discovery", "Response")
	}, "hub", "discovery")
	fmt.Println("Server: discovery handling started")
}

//Deprecated
type jsonobjectTestFile struct {
	Hubs []HubsType
}

//Deprecated
func __getHubList() []HubsType {
	//this function read json file
	file, err := ioutil.ReadFile("./ListHubs.json")
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}

	var jsontype jsonobjectTestFile
	err = json.Unmarshal(file, &jsontype)
	return jsontype.Hubs
}

func (srv *Server) GetPubKeyString() string {
	pkString := string(crypto.FromECDSAPub(&srv.PrivateKey.PublicKey))
	return pkString
}
