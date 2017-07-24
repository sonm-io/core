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

const Enode = "enode://b0605764bd7c6a816c51325a9cb9d414277d639f420f9dc48b20d12c04c33391b0a99cc8c045d7ba4657de0c04e8bb3b0d4b072ca9779167a75761d7c3c18eb0@10.196.131.151:30348"
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
