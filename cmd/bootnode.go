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
	"bufio"
	"crypto/ecdsa"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strings"

	"net/http"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/nat"
	whisper "github.com/ethereum/go-ethereum/whisper/whisperv2"
	"github.com/sonm-io/fusrodah/util"
)

const (
	quitCommand         = "~Q"
	httpInfoPort        = 8092
	httpInfoPath        = "/info"
	defaultBootnodePort = ":30348"
)

// singletons
var (
	server *p2p.Server
	shh    *whisper.Whisper
	done   chan struct{}

	input = bufio.NewReader(os.Stdin)
)

// encryption
var (
	pub    *ecdsa.PublicKey
	nodeid *ecdsa.PrivateKey
)

// cmd arguments
var (
	bootstrapMode = flag.Bool("standalone", true, "boostrap node: don't actively connect to peers, wait for incoming connections")
	generateKey   = flag.Bool("generatekey", false, "generate and show the private key")

	argVerbosity = flag.Int("verbosity", int(log.LvlError), "log verbosity level")
	argTTL       = flag.Uint("ttl", 30, "time-to-live for messages in seconds")
	argWorkTime  = flag.Uint("work", 5, "work time in seconds")

	argIP     = flag.String("ip", "", "IP address and port of this node (e.g. 127.0.0.1:30303)")
	argDBPath = flag.String("dbpath", "", "path to the server's DB directory")
	argIDFile = flag.String("idfile", "", "file name with node id (private key)")
	argEnode  = flag.String("boot", "", "bootstrap node you want to connect to (e.g. enode://e454......08d50@52.176.211.200:16428)")
)

func main() {
	initialize()
	echo()
	run()
}

func echo() {
	fmt.Printf("ttl = %d \n", *argTTL)
	fmt.Printf("workTime = %d \n", *argWorkTime)
	fmt.Printf("ip = %s \n", *argIP)
	fmt.Printf("pub = %s \n", common.ToHex(crypto.FromECDSAPub(pub)))
	fmt.Printf("idfile = %s \n", *argIDFile)
	fmt.Printf("dbpath = %s \n", *argDBPath)
	fmt.Printf("boot = %s \n", *argEnode)
	fmt.Printf("enode = %s \n", server.NodeInfo().Enode)
	fmt.Printf("http info = http://%s%s \n", getHttpInfoListenAddr(), httpInfoPath)
}

func initialize() {
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*argVerbosity), log.StreamHandler(os.Stderr, log.TerminalFormat(false))))

	done = make(chan struct{})
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
	nodeid = shh.NewIdentity()

	maxPeers := 80
	if *bootstrapMode {
		maxPeers = 800
	}

	server = &p2p.Server{
		Config: p2p.Config{
			PrivateKey:     nodeid,
			MaxPeers:       maxPeers,
			Name:           common.MakeName("wnode-bootnode", "2.0"),
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
		utils.Fatalf("Failed to start Whisper peer: %s.", err)
	}

	fmt.Println(server.NodeInfo().Enode)
	fmt.Println("Bootstrap Whisper node started")
}

func run() {
	startServer()
	defer server.Stop()
	shh.Start(server)
	defer shh.Stop()
	startHttpServer()

	sendLoop()
}

func sendLoop() {
	for {
		s := scanLine("")
		if s == quitCommand {
			fmt.Println("Quit command received")
			close(done)
			break
		}
	}
}

func scanLine(prompt string) string {
	if len(prompt) > 0 {
		fmt.Print(prompt)
	}
	txt, err := input.ReadString('\n')
	if err != nil {
		utils.Fatalf("input error: %s", err)
	}
	txt = strings.TrimRight(txt, "\n\r")
	return txt
}

func getHttpInfoListenAddr() string {
	return fmt.Sprintf("%s:%d", util.GetLocalIP(), httpInfoPort)
}

func startHttpServer() {
	http.HandleFunc(httpInfoPath, func(w http.ResponseWriter, r *http.Request) {
		body := fmt.Sprintf(`<h1>Node info</h1>
		<ul>
		<li>ttl = %d</li>
		<li>workTime = %d</li>
		<li>ip = %s</li>
		<li>pub = %s</li>
		<li>idfile = %s</li>
		<li>dbpath = %s</li>
		<li>boot = %s</li>
		<li>enode = %s</li>
		</ul>`, *argTTL, *argWorkTime, *argIP, common.ToHex(crypto.FromECDSAPub(pub)),
			*argIDFile, *argDBPath, *argEnode, server.NodeInfo().Enode)

		w.Write([]byte(body))
	})

	log.Info("Starting HTTP server", "addr", getHttpInfoListenAddr())
	http.ListenAndServe(getHttpInfoListenAddr(), nil)
}
