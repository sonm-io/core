package main

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
	"io/ioutil"
	"encoding/json"
)

type Fusrodah struct {
	prv           *ecdsa.PrivateKey
	cfg           p2p.Config
	p2pServer     p2p.Server
	whisperServer *whisperv2.Whisper

	p2pServerStatus     string
	whisperServerStatus string
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
		PrivateKey: fusrodah.prv,
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

	fusrodah.whisperServerStatus = "running"
}

func (fusrodah *Fusrodah) getTopics() []whisperv2.Topic {
	// NOTE for single topic use NewTopicFromString
	// NOTE whisperv2 is a package, shh - running whisper entity. Do not mess with that.
	// NOTE topics logic can be finded in whisperv2/topic.go
	// Topic represents a cryptographically secure, probabilistic partial
	// classifications of a message, determined as the first (left) 4 bytes of the
	// SHA3 hash of some arbitrary data given by the original author of the message.
	// TODO: get tuple of topic there
	topics := whisperv2.NewTopicsFromStrings("my", "message")
	return topics
}

func (fusrodah *Fusrodah) getFilterTopics() [][]whisperv2.Topic {
	// Creating new filters for a few topics.
	// NOTE more info about filters in /whisperv2/filters.go
	// TODO: get tuple of topic there
	topics := whisperv2.NewFilterTopicsFromStringsFlat("my", "message")
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
	msg.TTL = 3600000
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
		From:   fusrodah.prv, // Sign it
		Topics: topics,
	})
	if err != nil {
		fmt.Println("could not create whisper envelope:", err)
	}
	envelope.TTL = 4800000
	return envelope
}

func (fusrodah *Fusrodah) Send(message string) {

	// start whisper server, if it not running yet
	if fusrodah.whisperServerStatus != "running" {
		fusrodah.start()
	}

	// wrap source message to *whisper2.Message Entity
	whMessage := fusrodah.createMessage(message)

	// get possibly topics
	topics := fusrodah.getTopics()

	// wrap message to envelope, it needed to sending
	envelop := fusrodah.createEnvelop(whMessage, topics)

	if err := fusrodah.whisperServer.Send(envelop); err != nil {
		fmt.Println(err)
	} else {
		// this block actually for testing
		// NOTE: delete this block or wrap more
		fmt.Println("message sended")
	}
}

func (fusrodah Fusrodah) addHandling(cb func(msg *whisperv2.Message)) {
	// TODO: add topics
	// TODO: add return id

	// start whisper server, if it not running yet
	if fusrodah.whisperServerStatus != "running" {
		fusrodah.start()
	}

	// add watcher with any topics
	fusrodah.whisperServer.Watch(whisperv2.Filter{
		//	setting up filter
		//Topics: fusrodah.getFilterTopics(),
		//	setting up handler
		//	NOTE: parser and sotrting info in message should be inside this func
		Fn: cb,
	})

}


type HubsType struct {
	Id                  int
	Name                string
	TimeOfStart         int // TODO: cast to time.* Object
	AccountingPeriod    int
	Balance             float64
	MiddleSizeOfPayment float64
}

/**
 /--------TEST--------/
 THIS FUNCTION ALREADY FOR TEST
 */

type jsonobjectTestFile struct {
	Hubs   []HubsType
}

func __getHubList() []HubsType{
	file, err := ioutil.ReadFile("./ListHubs.json")
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}

	//m := new(Dispatch)
	//var m interface{}
	var jsontype jsonobjectTestFile
	json.Unmarshal(file, &jsontype)
	//fmt.Printf("Results: %v\n", jsontype)
	return jsontype.Hubs
}


func main() {
	//This is generate standart private key..(just private ket, NOT ethereum key struct.)
	//For generating ethereum key struct (with ethereum address etc) - use keystore.newKey
	// Private key is secp256k1
	fmt.Println(__getHubList())
	prv, _ := crypto.GenerateKey()

	// initialize Fusrodah with private key
	frd := Fusrodah{prv: prv}

	// you may start server manually
	frd.start()



	// NOTE: you previously need to setup filter
	//Watch for changing specified filter.
	//frd.addHandling(func(msg *whisperv2.Message) {
	//	fmt.Println("Recived message: ", string(msg.Payload))
	//})
	//
	//// any message test
	//frd.Send("test1")
	//frd.Send("test2")
	//frd.Send("test3")
	//frd.Send("test4")
	//frd.Send("test5")
	//frd.Send("test6")
	//frd.Send("test7")
	//frd.Send("test8")
	//frd.Send("test9")
	//frd.Send("test10")
	//frd.Send("test11")
	//frd.Send("test12")
	//frd.Send("test13")
	//frd.Send("test14")
	//frd.Send("test15")

	select {}
}
