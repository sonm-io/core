package miner

import (
	"encoding/json"
	"fmt"
	//"github.com/ethereum/go-ethereum/whisper/whisperv2"
	"io/ioutil"

	"crypto/ecdsa"
	"github.com/sonm-io/fusrodah/fusrodah"
	"github.com/sonm-io/fusrodah/hub"
	"github.com/ethereum/go-ethereum/crypto"
	//"time"
)

const Enode = "enode://e97d851aa39884a54320539f5dcab2ec6688e66116459e42b6d57c1d0db68107475875ad0d42230d97ee19a96440f7eba3f7273b8072d10afd4032e321a1f456@172.16.1.128:30348"
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
	ip         *string
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

}

func (srv *Server) GeHubIp() string {
	c := make(chan bool)
	if srv.ip == nil {
		go func() {
			srv.discovery()
			c <- true
		}()
	}

	//Filters here

	<-c
	return *srv.ip
}


func (srv *Server) GetPubKeyString() string {
	pkString := string(crypto.FromECDSAPub(&srv.PrivateKey.PublicKey))
	return pkString
}
