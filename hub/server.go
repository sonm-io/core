package hub

import (
	"github.com/sonm-io/go-ethereum/whisper/whisperv2"
	"fmt"
	"encoding/json"
	"github.com/sonm-io/Fusrodah"
	"io/ioutil"
	"os"
	"net"
)

/**
 /--------HUB--------/
 HUB FUNCTION SECTION
 /--------------------/
*/



type Server struct {
	//PrivateKey 	ecdsa.PrivateKey
	frd	Fusrodah.Fusrodah
	KnowingHubs []HubsType
	confFile    string
}

func hubMainFunction() {
	//TODO: need to feel main flow of the mainer in this function
}

func (hub *Server) loadKnowingHubs() {
	// NOTE: this for test case any
	hub.KnowingHubs = __getHubList()
}

func (hub *Server) DiscoveryHandling(frd Fusrodah.Fusrodah) {
	//this function load knowing hubs and at the same time
	//and print hubs with topics
	frd.AddHandling(func(msg *whisperv2.Message) {
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
		//fmt.Println("TESTTTTTTTTTT:", string(hubListString))
		frd.Send(string(hubListString), nil,  "hub", "discovery", "Response")
	}, "hub", "discovery")
	fmt.Println("Server: discovery handling started")

}




type jsonobjectTestFile struct {
	Hubs []HubsType
}


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