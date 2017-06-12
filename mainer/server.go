package mainer

import (
	"io/ioutil"
	"fmt"
	"encoding/json"
	"github.com/sonm-io/go-ethereum/whisper/whisperv2"

	"github.com/sonm-io/Fusrodah"
	"github.com/sonm-io/Fusrodah/hub"
)

/**
 /--------MAINER--------/
 MAINER FUNCTION SECTION
 /--------------------/
*/
type Server struct {
	//PrivateKey 	ecdsa.PrivateKey
	Hubs     []hub.HubsType
	ConfFile string
}

func mainerMainFunction() {

}

func (mainer *Server) LoadConf() bool {
	//this function load miners configuration
	file, err := ioutil.ReadFile(mainer.ConfFile)
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
	*mainer = m
	return true
}
func (mainer Server) SaveConf() bool {
	//this function save miners configuration
	hubListString, err := json.Marshal(mainer)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// NOTE: this for test
	fmt.Println("list:", string(hubListString))

	err = ioutil.WriteFile(mainer.ConfFile, hubListString, 0644)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func (mainer Server) StartDiscovery(frd Fusrodah.Fusrodah) bool{
	//now we send a message with topics
	verifyMsg := "{\"message\":\"verify\"}";
	//verifyMsg := "{"+'"'+"message"+'"'+":"+'"'+"verify"+'"'+"}"
	//json view {"message":"verify"}

	defer frd.Send(verifyMsg, nil,  "hub", "discovery")
	//Expect a response from the hub
	//which sends information about itself
	//with the topics "hub", "discovery", "Response"

	c := make(chan bool)

	go func() {
		frd.AddHandling(nil, func(msg *whisperv2.Message) {

			m := Server{}
			err := json.Unmarshal(msg.Payload, &m.Hubs)
			fmt.Println("Server: discoveryHand: ", m.Hubs)
			mainer.Hubs = m.Hubs
			if err != nil {
				fmt.Println(err)
				c <- false
			}
			fmt.Println("MAIN MAINER 2", mainer.Hubs)
			c <- true

		}, "hub", "discovery", "Response")
	}()

	<-c
	return true

}

func (mainer Server)GetAddress() string{

	return "success"
}
