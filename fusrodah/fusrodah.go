package fusrodah

/*
	This program use modified go-ethereum library (https://github.com/sonm-io/go-ethereum)
	Author Sonm.io team (@sonm-io on GitHub)
	Copyright 2017
*/

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/sonm-io/fusrodah/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/nat"
	"github.com/ethereum/go-ethereum/whisper/whisperv2"
	"os"
)

type Fusrodah struct {
	Prv           *ecdsa.PrivateKey
	cfg           p2p.Config
	p2pServer     p2p.Server
	whisperServer *whisperv2.Whisper

	p2pServerStatus     string
	whisperServerStatus string

	Enode string
	Port  string
}

func (fusrodah *Fusrodah) Start() {
	// function that start whisper server
	// private key is needed

	//Creates new instance of whisper protocol entity. NOTE - using whisper v.2 (not v5)
	fusrodah.whisperServer = whisperv2.New()

	//Configuration to running p2p server. Configuration values can't be modified after launch.
	//See p2p package in go-ethereum (server.go) for more info.
	//fusrodah.cfg = p2p.Config{
	//	MaxPeers: 10,
	//	//	Identity:   p2p.NewSimpleClientIdentity("my-whisper-app", "1.0", "", string(pub)),
	//	PrivateKey: fusrodah.Prv,
	//	ListenAddr: ":8000",
	//
	//	//here we can define what additional protocols will be used *above* p2p server.
	//	Protocols: []p2p.Protocol{whisperv2.Whisper{}.Protocol},
	//}

	var peers []*discover.Node
	peer := discover.MustParseNode(fusrodah.Enode)
	peers = append(peers, peer)

	maxPeers := 80

	tmpID := fusrodah.whisperServer.NewIdentity()

	//Definition of p2p server and binds to configuration. Configuration also could be stored in file.
	fusrodah.p2pServer = p2p.Server{
		Config: p2p.Config{
			PrivateKey:     tmpID,
			MaxPeers:       maxPeers,
			Name:           common.MakeName("wnode", "2.0"),
			Protocols:      fusrodah.whisperServer.Protocols(),
			ListenAddr:     util.GetLocalIP() + fusrodah.Port,
			NAT:            nat.Any(),
			BootstrapNodes: peers,
			StaticNodes:    peers,
			TrustedNodes:   peers,
		},
	}

	//Starting server and listen to errors.
	// TODO: experience with this
	// may trouble with starting p2p not needed exactly
	if err := fusrodah.p2pServer.Start(); err != nil {
		fmt.Println("could not start server:", err)
		os.Exit(1)
	}

	//Starting whisper protocol on running server.
	// NOTE whisper *should* be started automatically but it is not happening... possible BUG in go-ethereum.
	if err := fusrodah.whisperServer.Start(fusrodah.p2pServer); err != nil {
		fmt.Println("could not start server:", err)
		os.Exit(1)
	}

	fusrodah.whisperServerStatus = "running"
}

func (fusrodah *Fusrodah) Stop() {
	fusrodah.whisperServer.Stop()
	fusrodah.p2pServer.Stop()
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

func (fusrodah *Fusrodah) createMessage(message string, to *ecdsa.PublicKey) *whisperv2.Message {
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
	//TTL-hop limit is a mechanism that limits the lifespan or lifetime of message in a network
	msg.To = to
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
		From:   fusrodah.Prv, // Sign it
		Topics: topics,
	})
	if err != nil {
		fmt.Println("could not create whisper envelope:", err)
	}
	envelope.TTL = 4800000
	return envelope
}

func (fusrodah *Fusrodah) Send(message string, to *ecdsa.PublicKey, topics ...string) {

	// start whisper server, if it not running yet
	if fusrodah.whisperServerStatus != "running" {
		fusrodah.Start()
	}

	// wrap source message to *whisper2.Message Entity
	whMessage := fusrodah.createMessage(message, nil)

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

func (fusrodah *Fusrodah) AddHandling(to *ecdsa.PublicKey, cb func(msg *whisperv2.Message), topics ...string) int {
	// start whisper server, if it not running yet
	if fusrodah.whisperServerStatus != "running" {
		fusrodah.Start()
	}

	// add watcher with any topics
	id := fusrodah.whisperServer.Watch(whisperv2.Filter{
		//	setting up filter
		Topics: fusrodah.getFilterTopics(topics...),
		//	setting up handler
		//	NOTE: parser and sotrting info in message should be inside this func
		Fn: cb,
		To: to,
	})
	return id
}
