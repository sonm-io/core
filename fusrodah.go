package Fusrodah

/*
	This program use modified go-ethereum library (https://github.com/sonm-io/go-ethereum)
	Author Sonm.io team (@sonm-io on GitHub)
	Copyright 2017
*/

import (
	"bufio"
	"crypto/ecdsa"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/nat"
	whisper "github.com/ethereum/go-ethereum/whisper/whisperv5"
	"github.com/ethereum/go-ethereum/crypto"
)

const quitCommand = "~Q"

type FusrodahConfig struct {
	Verbosity int

	Enode string

	P2pAccept  bool
	SymKey     []byte
	Pub        *ecdsa.PublicKey
	AsymKey    *ecdsa.PrivateKey
	Nodeid     *ecdsa.PrivateKey
	Topic      whisper.TopicType
	AsymKeyID  string
	FilterID   string
	SymPass    string
	MsPassword string
}

var DefaultConfig = FusrodahConfig{
	Verbosity: 10,

	Enode: "enode://34adf2cbb2331336163c4e53ae666886fc40e072521d1ccbfe0041aacfa314d6c892f9ca295345e6f5fee1ebbe22b8b8173ca7c521c373eb19ede168c8af6452@172.16.1.10:30348",

	P2pAccept: true,

}

type Fusrodah struct {
	config 				*FusrodahConfig
	server 				*p2p.Server
	shh    				*whisper.Whisper
	done   				chan struct{}

	input 				*bufio.Reader

	whisperServerStatus bool
}

func NewFusrodah(config *FusrodahConfig) Fusrodah {
	if config == nil {
		config = &DefaultConfig
	}
	frd := Fusrodah{
		config: config,
		whisperServerStatus: false,
	}
	return frd
}

func (fusrodah *Fusrodah) Init() error {
	fusrodah.input = bufio.NewReader(os.Stdin)
	fusrodah.done = make(chan struct{})
	var err error
	var peers []*discover.Node

	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(fusrodah.config.Verbosity), log.StreamHandler(os.Stderr, log.TerminalFormat(false))))

	peer := discover.MustParseNode(fusrodah.config.Enode)
	peers = append(peers, peer)

	cfg := &whisper.Config{
		MaxMessageSize:     whisper.DefaultMaxMessageSize,
		MinimumAcceptedPOW: whisper.DefaultMinimumPoW,
	}

	fusrodah.shh = whisper.New(cfg)


	maxPeers := 80

	fusrodah.config.AsymKeyID, err = fusrodah.shh.NewKeyPair()
	if err != nil {
		utils.Fatalf("Failed to generate a new key pair: %s", err)
	}

	fusrodah.config.AsymKey, err = fusrodah.shh.GetPrivateKey(fusrodah.config.AsymKeyID)
	if err != nil {
		utils.Fatalf("Failed to retrieve a new key pair: %s", err)
	}


	symPass := "SOMN PASS"

	symKeyID, err := fusrodah.shh.AddSymKeyFromPassword(symPass)
	if err != nil {
		utils.Fatalf("Failed to create symmetric key: %s", err)
	}
	fusrodah.config.SymKey, err = fusrodah.shh.GetSymKey(symKeyID)
	if err != nil {
		utils.Fatalf("Failed to save symmetric key: %s", err)
	}

	if fusrodah.config.Nodeid == nil {
		tmpID, err := fusrodah.shh.NewKeyPair()
		if err != nil {
			utils.Fatalf("Failed to generate a new key pair: %s", err)
		}

		fusrodah.config.Nodeid, err = fusrodah.shh.GetPrivateKey(tmpID)
		if err != nil {
			utils.Fatalf("Failed to retrieve a new key pair: %s", err)
		}
	}

	fusrodah.server = &p2p.Server{
		Config: p2p.Config{
			PrivateKey:     fusrodah.config.Nodeid,
			MaxPeers:       maxPeers,
			Name:           common.MakeName("wnode", "5.0"),
			Protocols:      fusrodah.shh.Protocols(),
			ListenAddr:     GetLocalIP() + ":30349",
			NAT:            nat.Any(),
			BootstrapNodes: peers,
			StaticNodes:    peers,
			TrustedNodes:   peers,
		},
	}

	//fmt.Println("Started")

	return err
}

func (fusrodah *Fusrodah) Start() error {
	fusrodah.Init()

	err := fusrodah.server.Start()
	if err != nil {
		utils.Fatalf("Failed to start Whisper peer: %s.", err)
	}

	var cnt int
	var connected bool
	for !connected {
		//time.Sleep(time.Millisecond * 50)
		connected = fusrodah.server.PeerCount() > 0
		cnt++
		if cnt > 1000 {
			utils.Fatalf("Timeout expired, failed to connect")
		}
	}


	fmt.Println("Connected to peer.")


	//fusrodah.configureNode()

	return nil
}

func (fusrodah *Fusrodah) Send(message string, topic string) error {
	var err error

	//whTopic := whisper.BytesToTopic([]byte(topic))

	params := whisper.MessageParams{
		Src:     fusrodah.config.AsymKey,
		Dst:     fusrodah.config.Pub,
		KeySym:  fusrodah.config.SymKey,
		Payload: []byte(message),
		//Topic:   whTopic,
		TTL: 2000,
		PoW: 100000,
	}

	fmt.Println(fusrodah.config.SymKey)

	msg, err := whisper.NewSentMessage(&params)
	if err != nil {
		utils.Fatalf("failed to create new message: %s", err)
	}
	envelope, err := msg.Wrap(&params)
	if err != nil {
		fmt.Printf("failed to seal message: %v \n", err)
	}

	err = fusrodah.shh.Send(envelope)
	if err != nil {
		fmt.Printf("failed to send message: %v \n", err)
	} else {
		fmt.Println("message sended")
	}

	return err
}

func (fusrodah *Fusrodah) Stop() error {
	var err error
	return err
}

func (fusrodah *Fusrodah) AddHandler() error {
	var err error

	return err
}

func (fusrodah *Fusrodah) setFilter() error {
	var err error
	filter := whisper.Filter{
		KeySym:   fusrodah.config.SymKey,
		KeyAsym:  fusrodah.config.AsymKey,
		Topics:   [][]byte{fusrodah.config.Topic[:]},
		AllowP2P: fusrodah.config.P2pAccept,
	}
	filterID, err := fusrodah.shh.Subscribe(&filter)

	f := fusrodah.shh.GetFilter(filterID)
	if f == nil {
		return fmt.Errorf("filter is not installed")
	}

	ticker := time.NewTicker(time.Millisecond * 50)

	for {
		select {
		case <-ticker.C:
			messages := f.Retrieve()
			for _, msg := range messages {
				timestamp := fmt.Sprintf("%d", msg.Sent) // unix timestamp for diagnostics
				text := string(msg.Payload)

				var address common.Address
				if msg.Src != nil {
					address = crypto.PubkeyToAddress(*msg.Src)
				}

				if whisper.IsPubKeyEqual(msg.Src, &fusrodah.config.AsymKey.PublicKey) {
					fmt.Printf("\n%s <%x>: %s\n", timestamp, address, text) // message from myself
				} else {
					fmt.Printf("\n%s [%x]: %s\n", timestamp, address, text) // message from a peer
				}
			}
		}
	}

	return err

}

//
//import (
//	"fmt"
//	"os"
//	"crypto/ecdsa"
//	"github.com/sonm-io/go-ethereum/p2p"
//	"github.com/sonm-io/go-ethereum/whisper/whisperv2"
//	"github.com/sonm-io/go-ethereum/p2p/discover"
//)
//
//type Fusrodah struct {
//	Prv           *ecdsa.PrivateKey
//	cfg           p2p.Config
//	p2pServer     p2p.Server
//	whisperServer *whisperv2.Whisper
//
//	p2pServerStatus     string
//	whisperServerStatus string
//}
//
//var (
//	defaultPort = "8001"
//)
//
//func (fusrodah *Fusrodah) Start() {
//	// function that start whisper server
//	// private key is needed
//
//	//Creates new instance of whisper protocol entity. NOTE - using whisper v.2 (not v5)
//	fusrodah.whisperServer = whisperv2.New()
//
//	//Configuration to running p2p server. Configuration values can't be modified after launch.
//	//See p2p package in go-ethereum (server.go) for more info.
//	fusrodah.cfg = p2p.Config{
//		MaxPeers: 10,
//		//	Identity:   p2p.NewSimpleClientIdentity("my-whisper-app", "1.0", "", string(pub)),
//		PrivateKey: fusrodah.Prv,
//		ListenAddr: GetLocalIP()+":"+port,
//		//here we can define what additional protocols will be used *above* p2p server.
//		//Protocols: []p2p.Protocol{whisperv2.Whisper{}.Protocol},
//		//DiscoveryV5: true,
//	}
//
//	//Definition of p2p server and binds to configuration. Configuration also could be stored in file.
//	fusrodah.p2pServer = p2p.Server{
//		Config: fusrodah.cfg,
//	}
//
//	//Starting server and listen to errors.
//	// TODO: experience with this
//	// may trouble with starting p2p not needed exactly
//
//	if err := fusrodah.p2pServer.Start(); err != nil {
//		fmt.Println("could not start server:", err)
//		//	srv.Stop()
//		os.Exit(1)
//	}
//
//	//Starting whisper protocol on running server.
//	// NOTE whisper *should* be started automatically but it is not happening... possible BUG in go-ethereum.
//	if err := fusrodah.whisperServer.Start(fusrodah.p2pServer); err != nil {
//		fmt.Println("could not start server:", err)
//		//	srv.Stop()
//		os.Exit(1)
//	}
//	fusrodah.p2pServer.AddPeer(discover.MustParseNode("enode://75c3e18481f90709e85ee7e5883592f69003b71de5f8295b749b8469e29e57d8e779b314ea217cdea3938c2954f8b62506349e16a66ef6ea071c71398b1c0ba5@172.16.1.10:30348"))
//	//fusrodah.p2pServer.AddPeer(discover.MustParseNode("enode://618dec1c4df3c06bbba20f04886c39015257ff5039a32a744f9ab6485b610158c686f74ed8e901d5b83664946b136a880380f6e47578162869928e4212b93cc8@172.16.1.10:8001"))
//	fmt.Println(discover.NodeID{})
//	fmt.Println(fusrodah.p2pServer.Peers())
//
//	fmt.Println(fusrodah.p2pServer.DiscoveryV5)
//
//	fusrodah.whisperServerStatus = "running"
//}
//
//func (fusrodah *Fusrodah) Stop(){
//	fusrodah.whisperServer.Stop()
//	fusrodah.p2pServer.Stop()
//}
//
//func (fusrodah *Fusrodah) getTopics(data ...string) []whisperv2.Topic {
//	// NOTE for single topic use NewTopicFromString
//	// NOTE whisperv2 is a package, shh - running whisper entity. Do not mess with that.
//	// NOTE topics logic can be finded in whisperv2/topic.go
//	// Topic represents a cryptographically secure, probabilistic partial
//	// classifications of a message, determined as the first (left) 4 bytes of the
//	// SHA3 hash of some arbitrary data given by the original author of the message.
//	topics := whisperv2.NewTopicsFromStrings(data...)
//	return topics
//}
//
//func (fusrodah *Fusrodah) getFilterTopics(data ...string) [][]whisperv2.Topic {
//	// Creating new filters for a few topics.
//	// NOTE more info about filters in /whisperv2/filters.go
//	topics := whisperv2.NewFilterTopicsFromStringsFlat(data...)
//	return topics
//}
//
//func (fusrodah *Fusrodah) createMessage(message string, to *ecdsa.PublicKey) *whisperv2.Message {
//	// Creates entity of message itself.
//	// Message represents an end-user data packet to transmit through the Whisper
//	// protocol. These are wrapped into Envelopes that need not be understood by
//	// intermediate nodes, just forwarded.
//	/*
//	type Message struct {
//		Flags     byte // First bit is signature presence, rest reserved and should be random
//		Signature []byte
//		Payload   []byte
//
//		Sent time.Time     // Time when the message was posted into the network
//		TTL  time.Duration // Maximum time to live allowed for the message
//
//		To   *ecdsa.PublicKey // Message recipient (identity used to decode the message)
//		Hash common.Hash      // Message envelope hash to act as a unique id
//	}
//	*/
//	// NewMessage creates and initializes a non-signed, non-encrypted Whisper message.
//	// NOTE more info in whisperv2/message.go
//	// NOTE  first we create message, then we create envelope.
//	msg := whisperv2.NewMessage([]byte(message))
//	//TTL-hop limit is a mechanism that limits the lifespan or lifetime of message in a network
//	msg.To = to
//	msg.TTL = 3600000
//	return msg
//}
//
//func (fusrodah *Fusrodah) createEnvelop(message *whisperv2.Message, topics []whisperv2.Topic) *whisperv2.Envelope {
//	//Now we wrap message into envelope
//	// Wrap bundles the message into an Envelope to transmit over the network.
//	//
//	// pow (Proof Of Work) controls how much time to spend on hashing the message,
//	// inherently controlling its priority through the network (smaller hash, bigger
//	// priority).
//	//
//	// The user can control the amount of identity, privacy and encryption through
//	// the options parameter as follows:
//	//   - options.From == nil && options.To == nil: anonymous broadcast
//	//   - options.From != nil && options.To == nil: signed broadcast (known sender)
//	//   - options.From == nil && options.To != nil: encrypted anonymous message
//	//   - options.From != nil && options.To != nil: encrypted signed message
//	envelope, err := message.Wrap(1, whisperv2.Options{
//		From:   fusrodah.Prv, // Sign it
//		Topics: topics,
//		To: nil,
//	})
//	if err != nil {
//		fmt.Println("could not create whisper envelope:", err)
//	}
//	envelope.TTL = 4800000
//	return envelope
//}
//
//func (fusrodah *Fusrodah) Send(message string, to *ecdsa.PublicKey,  topics ...string) {
//
//	// start whisper server, if it not running yet
//	//if fusrodah.whisperServerStatus != "running" {
//	//	fusrodah.Start()
//	//}
//
//	// wrap source message to *whisper2.Message Entity
//	whMessage := fusrodah.createMessage(message, nil)
//
//	// get possibly topics
//	tops := fusrodah.getTopics(topics...)
//
//	// wrap message to envelope, it needed to sending
//	envelop := fusrodah.createEnvelop(whMessage, tops)
//
//	if err := fusrodah.whisperServer.Send(envelop); err != nil {
//		fmt.Println(err)
//	} else {
//		// this block actually for testing
//		// NOTE: delete this block or wrap more
//		fmt.Println("message sended")
//	}
//}
//
//
//func (fusrodah *Fusrodah) AddHandling(to *ecdsa.PublicKey, cb func(msg *whisperv2.Message), topics ...string) int {
//	// start whisper server, if it not running yet
//	//if fusrodah.whisperServerStatus != "running" {
//	//	fusrodah.Start()
//	//}
//
//	// add watcher with any topics
//	id := fusrodah.whisperServer.Watch(whisperv2.Filter{
//		//	setting up filter
//		Topics: fusrodah.getFilterTopics(topics...),
//		//	setting up handler
//		//	NOTE: parser and sotrting info in message should be inside this func
//		Fn: cb,
//		To: to,
//	})
//	return id
//}
