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

	"time"

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
	quitCommand         = "~Q"
	httpInfoPort        = 8092
	httpInfoPath        = "/info"
	defaultBootnodePort = ":30349"
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
	argBootNode  = flag.String("boot", "", "bootstrap node you want to connect to (e.g. enode://e454......08d50@52.176.211.200:16428)")
)

func main() {
	flag.Parse()
	initialize()
	echo()
	run()
}

func echo() {
	fmt.Printf("pub = %s \n", common.ToHex(crypto.FromECDSAPub(pub)))
	fmt.Printf("boot = %s \n", *argBootNode)
	fmt.Printf("enode = %s \n", server.NodeInfo().Enode)
	fmt.Printf("http info = http://%s%s \n", getHttpInfoListenAddr(), httpInfoPath)
}

func initialize() {
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*argVerbosity), log.StreamHandler(os.Stderr, log.TerminalFormat(false))))

	done = make(chan struct{})
	var peers []*discover.Node

	fmt.Println(*argBootNode)
	peer := discover.MustParseNode(*argBootNode)
	peers = append(peers, peer)

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
			Name:           common.MakeName("sonm-bootnode", "0.1"),
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
		s := scanLine(">>>")
		if s == quitCommand {
			fmt.Println("Quit command received")
			close(done)
			break
		}

		if server.PeerCount() == 0 {
			fmt.Println("No peers detected")
			close(done)
			break
		} else {
			fmt.Printf("Find %d peers, try to send payload=%s\r\n", server.PeerCount(), s)

			msg := whisper.NewMessage([]byte(s))
			env := whisper.NewEnvelope(3*time.Second, nil, msg)
			shh.Send(env)
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
		<li>ip = %s</li>
		<li>pub = %s</li>
		<li>boot = %s</li>
		<li>enode = %s</li>
		</ul>`, util.GetLocalIP(), common.ToHex(crypto.FromECDSAPub(pub)), *argBootNode, server.NodeInfo().Enode)

		w.Write([]byte(body))
	})

	log.Info("Starting HTTP server", "addr", getHttpInfoListenAddr())
	http.ListenAndServe(getHttpInfoListenAddr(), nil)
}
