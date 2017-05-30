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
type Mainer struct {
	//PrivateKey 	ecdsa.PrivateKey
	Hubs     []hub.HubsType
	ConfFile string
}

func mainerMainFunction() {

}

func (mainer *Mainer) LoadConf() bool {
	file, err := ioutil.ReadFile(mainer.ConfFile)
	if err != nil {
		fmt.Println(err)
		return false
	}

	var m Mainer
	err = json.Unmarshal(file, &m)
	if err != nil {
		fmt.Println(err)
		return false
	}
	*mainer = m
	return true
}
func (mainer Mainer) SaveConf() bool {
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

func (mainer Mainer) StartDiscovery(frd Fusrodah.Fusrodah) {
	//now we send a message with topics
	defer frd.Send("", "hub", "discovery")
	//Expect a response from the hub
	//which sends information about itself
	//with the topics "hub", "discovery", "Response"
	frd.AddHandling(func(msg *whisperv2.Message) {
		m := Mainer{}
		//fmt.Println(string(msg.Payload))
		err := json.Unmarshal(msg.Payload, &m.Hubs)
		fmt.Println("Mainer: discoveryHand: ", m.Hubs)
		mainer.Hubs = m.Hubs
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("MAIN MAINER 2", mainer.Hubs)
		defer mainer.firstFilter(2.4)
		defer mainer.secondFilter(10)
		defer mainer.AccountingPeriodFilter(3)

	}, "hub", "discovery", "Response")
}
func (mainer Mainer) firstFilter(neededBalance float64) []hub.HubsType {

	//use filter: balance more then neededBalance
	var someList []hub.HubsType
	for _, hub := range mainer.Hubs {
		if hub.Balance >= neededBalance {
			someList = append(someList, hub)
		}
	}

	mainer.Hubs = someList
	fmt.Println("WHITELIST", mainer.Hubs)
	return someList
}
func (mainer Mainer) secondFilter(neededBalance float64) []hub.HubsType {
	//use filter: balance less then neededBalance
	var someList []hub.HubsType
	for _, hub := range mainer.Hubs {
		if hub.Balance <= neededBalance {
			someList = append(someList, hub)
		}
	}
	mainer.Hubs = someList
	fmt.Println("WhiteList2", mainer.Hubs)
	return someList
}
func (mainer Mainer) AccountingPeriodFilter(neededPeriod int) []hub.HubsType {
	//use filter: accountingPeriod > neededPeriod
	var someList []HubsType
	for _, hub := range mainer.Hubs {
		if hub.AccountingPeriod > neededPeriod {
			someList = append(someList, hub)
		}
	}
	mainer.Hubs = someList
	fmt.Println("FilterPeriodList", mainer.Hubs)
	return someList
}
