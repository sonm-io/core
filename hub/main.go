package hub

import ()
import (
	"github.com/sonm-io/go-ethereum/whisper/whisperv2"
	"fmt"
	"encoding/json"
	"github.com/sonm-io/Fusrodah"
)

/**
 /--------HUB--------/
 HUB FUNCTION SECTION
 /--------------------/
*/



type Hub struct {
	//PrivateKey 	ecdsa.PrivateKey
	frd	Fusrodah.Fusrodah
	KnowingHubs []HubsType
	confFile    string
}

func hubMainFunction() {
	//TODO: need to feel main flow of the mainer in this function
}

func (hub *Hub) loadKnowingHubs() {
	// NOTE: this for test case any
	hub.KnowingHubs = __getHubList()
}

func (hub *Hub) DiscoveryHandling(frd Fusrodah.Fusrodah) {
	frd.AddHandling(func(msg *whisperv2.Message) {
		hub.loadKnowingHubs()
		fmt.Println("Hub: discovery response")
		hubListString, err := json.Marshal(hub.KnowingHubs)
		if err != nil {
			fmt.Println(err)
			return
		}
		//fmt.Println("TESTTTTTTTTTT:", string(hubListString))
		frd.Send(string(hubListString), "hub", "discovery", "Response")
	}, "hub", "discovery")
	fmt.Println("Hub: discovery handling started")

}
