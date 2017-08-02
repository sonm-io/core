package fusrodah

/*
	This program use modified go-ethereum library (https://github.com/sonm-io/go-ethereum)
	Author Sonm.io team (@sonm-io on GitHub)
	Copyright 2017
*/

import (
	"crypto/ecdsa"
	"errors"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/nat"
	"github.com/ethereum/go-ethereum/whisper/whisperv2"

	frdConst "github.com/sonm-io/core/common"
	"github.com/sonm-io/core/util"
)

type serverState int

const (
	serverStateStopped = 0
	serverStateRunning = 1
	maxPeers           = 80
)

var (
	errServerNotRunning = errors.New("Server is not running")
)

type Fusrodah struct {
	Prv *ecdsa.PrivateKey

	p2pServer           p2p.Server
	whisperServer       *whisperv2.Whisper
	whisperServerStatus serverState

	Enodes []string
	Port   string
}

// NewServer builds new Fusrodah server instance
func NewServer(prv *ecdsa.PrivateKey, port string, enodes []string) (frd *Fusrodah, err error) {

	if prv == nil {
		prv, err = crypto.GenerateKey()
		if err != nil {
			return nil, err
		}
	}

	shh := whisperv2.New()

	frd = &Fusrodah{
		Prv:                 prv,
		Port:                port,
		Enodes:              enodes,
		whisperServer:       shh,
		whisperServerStatus: serverStateStopped,
	}
	return frd, nil
}

// Start start whisper server
func (fusrodah *Fusrodah) Start() (err error) {
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(frdConst.DefaultLogLevelGoEthereum), log.StreamHandler(os.Stderr, log.TerminalFormat(false))))
	// Creates new instance of whisper protocol entity. NOTE - using whisper v.2 (not v5)
	var peers []*discover.Node

	for _, enode := range fusrodah.Enodes {
		peer := discover.MustParseNode(enode)
		peers = append(peers, peer)
	}

	// Configuration to running p2p server. Configuration values can't be modified after launch.
	fusrodah.p2pServer = p2p.Server{
		Config: p2p.Config{
			PrivateKey:     fusrodah.Prv,
			MaxPeers:       maxPeers,
			Name:           common.MakeName("wpeer", "2.0"),
			Protocols:      fusrodah.whisperServer.Protocols(),
			ListenAddr:     util.GetLocalIP() + fusrodah.Port,
			NAT:            nat.Any(),
			BootstrapNodes: peers,
			StaticNodes:    peers,
			TrustedNodes:   peers,
		},
	}

	// Starting p2p server
	err = fusrodah.p2pServer.Start()
	if err != nil {
		return err
	}

	// Starting whisper protocol on running p2p server.
	// NOTE whisper *should* be started automatically but it is not happening... possible BUG in go-ethereum.
	err = fusrodah.whisperServer.Start(&fusrodah.p2pServer)
	if err != nil {
		return err
	}

	//log.Info("my public key", "key", common.ToHex(crypto.FromECDSAPub(&fusrodah.asymKey.PublicKey)))
	fusrodah.whisperServerStatus = serverStateRunning
	return nil
}

// Stop stops whisper and p2p servers
func (fusrodah *Fusrodah) Stop() (err error) {
	err = fusrodah.whisperServer.Stop()
	if err != nil {
		return err
	}
	fusrodah.p2pServer.Stop()
	return nil
}

// getFilterTopics Creating new filters for a few topics.
// NOTE more info about filters in /whisperv2/filters.go
func (fusrodah *Fusrodah) getFilterTopics(data ...string) [][]whisperv2.Topic {
	topics := whisperv2.NewFilterTopicsFromStringsFlat(data...)
	return topics
}

// isRunning check if Fusrodah server is running
func (fusrodah *Fusrodah) isRunning() bool {
	return fusrodah.whisperServerStatus == serverStateRunning
}

// Send sends broadcast send non-encrypted message
func (fusrodah *Fusrodah) Send(payload string, anonymous bool, topics ...string) error {
	if !fusrodah.isRunning() {
		return errServerNotRunning
	}

	var from *ecdsa.PrivateKey
	if anonymous {
		from = nil
	} else {
		from = fusrodah.Prv
	}

	opts := whisperv2.Options{
		From:   from,
		To:     nil,
		Topics: whisperv2.NewTopicsFromStrings(topics...),
		TTL:    whisperv2.DefaultTTL,
	}

	msg := whisperv2.NewMessage([]byte(payload))
	env, err := msg.Wrap(whisperv2.DefaultPoW, opts)
	if err != nil {
		return err
	}

	err = fusrodah.whisperServer.Send(env)
	if err != nil {
		return err
	}

	return nil
}

// AddHandling adds register handler for messages with given keys and on given topics
func (fusrodah *Fusrodah) AddHandling(to *ecdsa.PublicKey, from *ecdsa.PublicKey, handler func(msg *whisperv2.Message), topics ...string) int {
	if !fusrodah.isRunning() {
		fusrodah.Start()
	}

	// add watcher with any topics
	id := fusrodah.whisperServer.Watch(whisperv2.Filter{
		// setting up filter by topic
		Topics: fusrodah.getFilterTopics(topics...),
		// setting up message handler
		Fn: handler,
		// settings up sender and recipient
		From: from,
		To:   to,
	})

	return id
}

// RemoveHandling removes message handler by their id
func (fusrodah *Fusrodah) RemoveHandling(id int) {
	fusrodah.whisperServer.Unwatch(id)
}
