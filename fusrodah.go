package Fusrodah

/*
	This program use modified go-ethereum library (https://github.com/sonm-io/go-ethereum)
	Author Sonm.io team (@sonm-io on GitHub)
	Copyright 2017
*/

import (
	"fmt"
	"os"
	"crypto/ecdsa"
	"github.com/sonm-io/go-ethereum/p2p"
	"github.com/sonm-io/go-ethereum/crypto"
	"github.com/sonm-io/go-ethereum/whisper/whisperv2"
	"github.com/sonm-io/Fusrodah/hub"
	"github.com/sonm-io/Fusrodah/mainer"
	"io/ioutil"
	"encoding/json"
)


// Time To Live - relation to message in Whisper protocol
const TTL = 3600 * 2


/**
  	SONM wrapper to whisper protocol
  	@param Prv - private key to identify sender/
  	@param cfg - instance of p2p.Config object
  	@param p2pServer - instance of p2p.Server
  	@param whisperServer - instance of whisperv2.Whisper - main relation instance

  	@param StatusP2pServer - string param that semaphore status of p2pServer
  	@param StatusWhisperServer - string param that semaphore status of whisper Server, correctly state is "running"
 */
type Fusrodah struct {
	Prv           *ecdsa.PrivateKey
	cfg           p2p.Config
	p2pServer     p2p.Server
	whisperServer *whisperv2.Whisper

	StatusP2pServer     string
	StatusWhisperServer string
}

func (fusrodah *Fusrodah) start() {
	// function that start whisper server
	// private key is needed

	//Creates new instance of whisper protocol entity. NOTE - using whisper v.2 (not v5)
	fusrodah.whisperServer = whisperv2.New()

	//Configuration to running p2p server. Configuration values can't be modified after launch.
	//See p2p package in go-ethereum (server.go) for more info.
	fusrodah.cfg = p2p.Config{
		MaxPeers: 10,
		//	Identity:   p2p.NewSimpleClientIdentity("my-whisper-app", "1.0", "", string(pub)),
		PrivateKey: fusrodah.Prv,
		ListenAddr: ":8000",

		//here we can define what additional protocols will be used *above* p2p server.
		Protocols: []p2p.Protocol{whisperv2.Whisper{}.Protocol},
	}

	//Definition of p2p server and binds to configuration. Configuration also could be stored in file.
	fusrodah.p2pServer = p2p.Server{
		Config: fusrodah.cfg,
	}

	//Starting server and listen to errors.
	// TODO: experience with this
	// may trouble with starting p2p not needed exactly
	//if err := fusrodah.p2pServer.Start(); err != nil {
	//	fmt.Println("could not start server:", err)
	//	//	srv.Stop()
	//	os.Exit(1)
	//}

	//Starting whisper protocol on running server.
	// NOTE whisper *should* be started automatically but it is not happening... possible BUG in go-ethereum.
	if err := fusrodah.whisperServer.Start(fusrodah.p2pServer); err != nil {
		fmt.Println("could not start server:", err)
		//	srv.Stop()
		os.Exit(1)
	}

	fusrodah.StatusWhisperServer = "running"
}

func (fusrodah *Fusrodah) getTopics(data ...string) []whisperv2.Topic {
	// NOTE for single topic use NewTopicFromString
	// NOTE whisperv2 is a package, shh - running whisper entity. Do not mess with that.
	// NOTE topics logic can be finded in whisperv2/topic.go
	// Topic represents a cryptographically secure, probabilistic partial
	// classifications of a message, determined as the first (left) 4 bytes of the
	// SHA3 hash of some arbitrary data given by the original author of the message.
	topics := whisperv2.NewTopicsFromStrings(data...)
	return topics
}

func (fusrodah *Fusrodah) getFilterTopics(data ...string) [][]whisperv2.Topic {
	// Creating new filters for a few topics.
	// NOTE more info about filters in /whisperv2/filters.go
	topics := whisperv2.NewFilterTopicsFromStringsFlat(data...)
	return topics
}

func (fusrodah *Fusrodah) createMessage(message string) *whisperv2.Message {
	// Creates entity of message itself.
	// Message represents an end-user data packet to transmit through the Whisper
	// protocol. These are wrapped into Envelopes that need not be understood by
	// intermediate nodes, just forwarded.
	/*
	type Message struct {
		Flags     byte // First bit is signature presence, rest reserved and should be random
		Signature []byte
		Payload   []byte

		Sent time.Time     // Time when the message was posted into the network
		TTL  time.Duration // Maximum time to live allowed for the message

		To   *ecdsa.PublicKey // Message recipient (identity used to decode the message)
		Hash common.Hash      // Message envelope hash to act as a unique id
	}
	*/
	// NewMessage creates and initializes a non-signed, non-encrypted Whisper message.
	// NOTE more info in whisperv2/message.go
	// NOTE  first we create message, then we create envelope.
	msg := whisperv2.NewMessage([]byte(message))
	msg.TTL = TTL
	return msg
}

func (fusrodah *Fusrodah) createEnvelop(message *whisperv2.Message, topics []whisperv2.Topic) *whisperv2.Envelope {
	//Now we wrap message into envelope
	// Wrap bundles the message into an Envelope to transmit over the network.
	//
	// pow (Proof Of Work) controls how much time to spend on hashing the message,
	// inherently controlling its priority through the network (smaller hash, bigger
	// priority).
	//
	// The user can control the amount of identity, privacy and encryption through
	// the options parameter as follows:
	//   - options.From == nil && options.To == nil: anonymous broadcast
	//   - options.From != nil && options.To == nil: signed broadcast (known sender)
	//   - options.From == nil && options.To != nil: encrypted anonymous message
	//   - options.From != nil && options.To != nil: encrypted signed message
	envelope, err := message.Wrap(1, whisperv2.Options{
		From:   fusrodah.Prv, // Sign it
		Topics: topics,
	})
	if err != nil {
		fmt.Println("could not create whisper envelope:", err)
	}
	envelope.TTL = TTL
	return envelope
}

func (fusrodah *Fusrodah) Send(message string, topics ...string) {

	// start whisper server, if it not running yet
	if fusrodah.StatusWhisperServer != "running" {
		fusrodah.start()
	}

	// wrap source message to *whisper2.Message Entity
	whMessage := fusrodah.createMessage(message)

	// get possibly topics
	tops := fusrodah.getTopics(topics...)

	// wrap message to envelope, it needed to sending
	envelop := fusrodah.createEnvelop(whMessage, tops)

	if err := fusrodah.whisperServer.Send(envelop); err != nil {
		fmt.Println(err)
	} else {
		// this block actually for testing
		// NOTE: delete this block or wrap more
		fmt.Println("message sended")
	}
}

func (fusrodah *Fusrodah) AddHandling(cb func(msg *whisperv2.Message), topics ...string) int {
	// start whisper server, if it not running yet
	if fusrodah.StatusWhisperServer != "running" {
		fusrodah.start()
	}

	// add watcher with any topics
	id := fusrodah.whisperServer.Watch(whisperv2.Filter{
		//	setting up filter
		Topics: fusrodah.getFilterTopics(topics...),
		//	setting up handler
		//	NOTE: parser and sotrting info in message should be inside this func
		Fn: cb,
	})
	return id
}



/**
 /--------TEST--------/
 THIS FUNCTION FOR TEST
 /--------------------/
*/

type jsonobjectTestFile struct {
	Hubs []hub.HubsType
}


func __getHubList() []hub.HubsType {
	file, err := ioutil.ReadFile("./ListHubs.json")
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}

	var jsontype jsonobjectTestFile
	err = json.Unmarshal(file, &jsontype)
	return jsontype.Hubs
}


func testsFn() {

	// test save configuration
	hubList := __getHubList()
	mainer1 := mainer.Mainer{ConfFile: "mainerConf.json"}
	//mainer.PrivateKey = *prv
	mainer1.Hubs = hubList
	mainer1.SaveConf()

	// test loading configuration
	mainer2 := mainer.Mainer{ConfFile: "mainerConf.json"}
	mainer2.LoadConf()
	mainer2.ConfFile = "mainerConf2.json"
	mainer2.SaveConf()
	fmt.Println(mainer2)

	//This is generate standart private key..(just private ket, NOT ethereum key struct.)
	//For generating ethereum key struct (with ethereum address etc) - use keystore.newKey
	// Private key is secp256k1
	prv, _ := crypto.GenerateKey()
	// initialize Fusrodah with private key
	frd := Fusrodah{Prv: prv}

	// you may start server manually
	frd.start()

	// NOTE: you previously need to setup filter
	//Watch for changing specified filter.
	handleId := frd.AddHandling(func(msg *whisperv2.Message) {
		fmt.Println("Recived message: ", string(msg.Payload))
	}, "test")

	fmt.Println("HandleID:", handleId)

	// any message test
	frd.Send("test1", "test")
	frd.Send("test2")
	frd.Send("test3")
	frd.Send("test4")
	frd.Send("test5")
	frd.Send("test6")
	frd.Send("test7")
	frd.Send("test8")
	frd.Send("test9")
	frd.Send("test10")
	frd.Send("test11")
	frd.Send("test12")
	frd.Send("test13")
	frd.Send("test14")
	frd.Send("test15")
}

func main() {

	/**
	HUB example
	 */
	hubPrv, _ := crypto.GenerateKey()
	hubFrd := Fusrodah{Prv: hubPrv}
	hubFrd.start()
	hub1 := hub.Hub{}
	hub1.DiscoveryHandling(hubFrd)

	/**
	Mainer example
	 */
	//mainer_1Prv, _ := crypto.GenerateKey()
	//mainer_1Frd := Fusrodah{prv: mainer_1Prv}
	//mainer_1Frd.start()
	mainer_1 := mainer.Mainer{}
	mainer_1.StartDiscovery(hubFrd)

	fmt.Println("MAIN MAINER 1", mainer_1.Hubs)

	/**
	any Mainer example
	 */
	//mainer_2Prv, _ := crypto.GenerateKey()
	//mainer_2Frd := Fusrodah{prv: mainer_2Prv}
	mainer_2 := mainer.Mainer{}
	mainer_2.StartDiscovery(hubFrd)
	select {}
}


