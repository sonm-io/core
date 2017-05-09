package main

import (
	"fmt"
	//"log"
	"os"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/whisper/whisperv5"
	//"github.com/ethereum/go-ethereum/crypto/secp256k1"
	//"github.com/ethereum/go-ethereum/crypto/secp256k1_test.go"
	"github.com/ethereum/go-ethereum/crypto"


)

func main() {




	prv, _ :=crypto.GenerateKey()



	shh := whisperv5.New()



	cfg := p2p.Config{
		MaxPeers:   10,
	//	Identity:   p2p.NewSimpleClientIdentity("my-whisper-app", "1.0", "", string(pub)),
		PrivateKey: prv,
		ListenAddr: ":8000",
	//	Protocols: []p2p.Protocol{whisper.protocol()},
		//	Protocols: []p2p.Protocol{shh.protocol()},
	//	Protocols:  []p2p.Protocol{MyProtocol()},
	Protocols: []p2p.Protocol{shh.Protocol},
	}



	srv:= p2p.Server{
		Config: cfg,
	}



  //srv.Start();


	if err := srv.Start(); err != nil {
		fmt.Println("could not start server:", err)
	//	srv.Stop()
		os.Exit(1)
	}



	select {}
}
