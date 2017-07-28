// Copyright 2017 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

// This is a simple Whisper node. It could be used as a stand-alone bootstrap node.
// Also, could be used for different test and diagnostics purposes.

package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"flag"
	"fmt"
	"os"

	"net/http"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/nat"
	whisper "github.com/ethereum/go-ethereum/whisper/whisperv2"
	"github.com/sonm-io/core/util"
)

const (
	httpInfoPort        = 8092
	httpInfoPath        = "/info"
	defaultBootnodePort = ":30348"
	applicationName     = "sonm-bootnode"
)

// singletons
var (
	server            *p2p.Server
	shh               *whisper.Whisper
	version           string
	obtainedEnodeAddr string
)

// encryption
var (
	pub *ecdsa.PublicKey
)

// cmd arguments
var (
	bootstrapMode = flag.Bool("standalone", true, "boostrap node: don't actively connect to peers, wait for incoming connections")
	generateKey   = flag.Bool("generatekey", false, "generate and show the private key")

	argVerbosity = flag.Int("verbosity", int(log.LvlInfo), "log verbosity level")
)

func main() {
	initialize()
	run()
	showServerInfo()

	select {}
}

func showServerInfo() {
	fmt.Printf("version   = %s \n", version)
	fmt.Printf("pub key   = %s \n", common.ToHex(crypto.FromECDSAPub(pub)))
	fmt.Printf("enode     = %s \n", obtainedEnodeAddr)
	fmt.Printf("http info = http://%s%s \n", getHttpInfoListenAddr(), httpInfoPath)
}

func initialize() {
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*argVerbosity), log.StreamHandler(os.Stderr, log.TerminalFormat(false))))

	var peers []*discover.Node
	if *generateKey {
		key, err := crypto.GenerateKey()
		if err != nil {
			utils.Fatalf("Failed to generate private key: %s", err)
		}
		k := hex.EncodeToString(crypto.FromECDSA(key))
		fmt.Printf("Random private key: %s \n", k)
		os.Exit(0)
	}

	localAddr := util.GetLocalIP() + defaultBootnodePort
	shh = whisper.New()

	maxPeers := 80
	if *bootstrapMode {
		maxPeers = 800
	}

	server = &p2p.Server{
		Config: p2p.Config{
			PrivateKey:     shh.NewIdentity(),
			MaxPeers:       maxPeers,
			Name:           common.MakeName(applicationName, version),
			Protocols:      shh.Protocols(),
			ListenAddr:     localAddr,
			NAT:            nat.Any(),
			BootstrapNodes: peers,
			StaticNodes:    peers,
			TrustedNodes:   peers,
		},
	}
}

func startServer() {
	err := server.Start()
	if err != nil {
		// todo(sshaman1101): handle error, break bootstrapping process if any
		utils.Fatalf("Failed to start Whisper peer: %s.", err)
	}

	obtainedEnodeAddr = server.NodeInfo().Enode
	log.Info("Bootstrap Node node started")
}

func run() {
	startServer()
	defer server.Stop()
	log.Info("Server started")

	shh.Start(server)
	log.Info("Whisper started")

	startHttpServer()
	log.Info("HTTP server started")
}

func getHttpInfoListenAddr() string {
	return fmt.Sprintf("%s:%d", util.GetLocalIP(), httpInfoPort)
}

func startHttpServer() {
	http.HandleFunc(httpInfoPath, func(w http.ResponseWriter, r *http.Request) {
		body := fmt.Sprintf(`<h1>Node info</h1>
		<ul>
		<li>version = %s</li>
		<li>ip addr = %s</li>
		<li>pub key = %s</li>
		<li>boot = %s</li>
		</ul>`, version, util.GetLocalIP(), common.ToHex(crypto.FromECDSAPub(pub)), obtainedEnodeAddr)

		w.Write([]byte(body))
	})

	log.Info("Starting HTTP server", "addr", getHttpInfoListenAddr())

	go func() {
		http.ListenAndServe(getHttpInfoListenAddr(), nil)
	}()
}
