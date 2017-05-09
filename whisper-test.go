package main

import (
	"fmt"
	//"log"
	"os"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/whisper/whisperv5"
	//"github.com/ethereum/go-ethereum/crypto/secp256k1"
	//"github.com/ethereum/go-ethereum/crypto/secp256k1_test.go"
	//"github.com/ethereum/go-ethereum/crypto"

	//"crypto/ecdsa"
	//"crypto/elliptic"
	//"crypto/rand"
	//"encoding/hex"
	//"gopkg.in/fatih/set.v0"
)

func main() {

	//pub, _ := secp256k1.generateKeyPair()

	//pub=generateKeyPair();

	//pub, _ :=crypto.ToECDSAPub(key)

//	prv, _ :=crypto.GenerateKey()

/**	func generateKeyPair() (pubkey, privkey []byte) {
		key, err := ecdsa.GenerateKey(S256(), rand.Reader)
		if err != nil {
			panic(err)
		}
		pubkey = elliptic.Marshal(S256(), key.X, key.Y)
		return pubkey, math.PaddedBigBytes(key.D, 32)
	}
**/

	shh := whisperv5.New()

	//shh.Protocol=shh.protocol

	cfg := p2p.Config{
		MaxPeers:   10,
	//	Identity:   p2p.NewSimpleClientIdentity("my-whisper-app", "1.0", "", string(pub)),
	//	PrivateKey: prv,
		ListenAddr: ":8000",
	//	Protocols: []p2p.Protocol{whisper.protocol()},
		//	Protocols: []p2p.Protocol{shh.protocol()},
	//	Protocols:  []p2p.Protocol{MyProtocol()},
	Protocols: []p2p.Protocol{shh.Protocol},
	}

 //ihatecompilers  := prv;

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
