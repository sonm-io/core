package miner

import (
	"io/ioutil"
	"fmt"
	"encoding/json"
	"github.com/sonm-io/go-ethereum/whisper/whisperv2"

	"github.com/sonm-io/Fusrodah/fusrodah"
	"github.com/sonm-io/Fusrodah/hub"
	"crypto/ecdsa"
	"github.com/sonm-io/go-ethereum/crypto"
	"time"
)

const Enode = "enode://81b8db7b071b46bfc8619268606df7edf48cc55f804f52ce6176bbb369cab22af752ce15c622c958f29dd7617c3d1d647f544f93ce5a11f4319334c418340e3c@172.16.1.111:30348"
const DEFAULT_MINER_PORT = ":30347"

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
	ip         *string
}

func NewServer(prv *ecdsa.PrivateKey) *Server {
	if prv == nil {
		//TODO: cover error
		prv, _ = crypto.GenerateKey()
	}

	frd := fusrodah.Fusrodah{
		Prv: prv,
		Enode: Enode,
		Port: DEFAULT_MINER_PORT,
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

//Deprecated
func (srv *Server) LoadConf() bool {
	//this function load miners configuration
	file, err := ioutil.ReadFile(srv.ConfFile)
	if err != nil {
		fmt.Println(err)
		return false
	}

	var m Server
	err = json.Unmarshal(file, &m)
	if err != nil {
		fmt.Println(err)
		return false
	}
	*srv = m
	return true
}

//Deprecated
func (srv *Server) SaveConf() bool {
	//this function save miners configuration
	hubListString, err := json.Marshal(srv)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// NOTE: this for test
	fmt.Println("list:", string(hubListString))

	err = ioutil.WriteFile(srv.ConfFile, hubListString, 0644)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func (srv *Server) discovery() {
	var hubPubKeyString *ecdsa.PublicKey
	c := make(chan bool, 1)

	go func() {

		srv.Frd.AddHandling(nil, func(msg *whisperv2.Message) {
			hubPubKeyString = crypto.ToECDSAPub(msg.Payload)
			c <- true
		}, "miner", "discover")

		for{
			srv.Frd.Send(srv.GetPubKeyString(), nil, "hubDiscover")
			fmt.Println("DISC #1 SENDED")
			time.Sleep(time.Millisecond * 1000)
		}
	}()

	<-c

	go func() {

		defer srv.Frd.Send(srv.GetPubKeyString(), hubPubKeyString, "hub", "addr")
		srv.Frd.AddHandling(&srv.PrivateKey.PublicKey, func(msg *whisperv2.Message) {
			*srv.ip = string(msg.Payload)
		}, "miner", "addr")
		c <- true
	}()

	<-c

}

func (srv *Server) GeHubIp() string {
	c := make(chan bool)
	if srv.ip == nil {
		go func(){
			srv.discovery()
			c <- true
		}()
	}

	//Filters here

	<-c
	return *srv.ip
}

func (srv *Server) StartDiscovery(frd fusrodah.Fusrodah) bool {
	//now we send a message with topics
	verifyMsg := "{\"message\":\"verify\"}";
	//verifyMsg := "{"+'"'+"message"+'"'+":"+'"'+"verify"+'"'+"}"
	//json view {"message":"verify"}

	defer frd.Send(verifyMsg, nil, "hub", "discovery")
	//Expect a response from the hub
	//which sends information about itself
	//with the topics "hub", "discovery", "Response"

	c := make(chan bool)

	go func() {
		frd.AddHandling(nil, func(msg *whisperv2.Message) {

			m := Server{}
			err := json.Unmarshal(msg.Payload, &m.Hubs)
			fmt.Println("Server: discoveryHand: ", m.Hubs)
			srv.Hubs = m.Hubs
			if err != nil {
				fmt.Println(err)
				c <- false
			}
			fmt.Println("MAIN MAINER 2", srv.Hubs)
			c <- true

		}, "hub", "discovery", "Response")
	}()

	<-c
	return true

}

func (srv *Server) GetPubKeyString() string {
	pkString := string(crypto.FromECDSAPub(&srv.PrivateKey.PublicKey))
	return pkString
}
