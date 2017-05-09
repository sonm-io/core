package main

import (
	"fmt"
	//"log"
	"os"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/whisper/whisperv2"
	//"github.com/ethereum/go-ethereum/crypto/secp256k1"
	//"github.com/ethereum/go-ethereum/crypto/secp256k1_test.go"
	"github.com/ethereum/go-ethereum/crypto"


)

func main() {




	prv, _ :=crypto.GenerateKey()



	shh := whisperv2.New()







	cfg := p2p.Config{
		MaxPeers:   10,
	//	Identity:   p2p.NewSimpleClientIdentity("my-whisper-app", "1.0", "", string(pub)),
		PrivateKey: prv,
		ListenAddr: ":8000",

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


	if err:=	shh.Start(srv); err != nil {
		fmt.Println("could not start whisper:", err)
	//	srv.Stop()
		os.Exit(1)
	}


topics := whisperv2.NewTopicsFromStrings("my", "message")
fmt.Println("topics")
fmt.Println(topics)
fmt.Println("test2")
//topics := whisperv2.TopicsFromString("my", "message")
//msg := shh.NewMessage([]byte("hello world"))  // 1
msg := whisperv2.NewMessage([]byte("hello world"))
fmt.Println("msg:")
fmt.Println(msg)

envelope, err := msg.Wrap(1,whisperv2.Options{                // 2
				From:   prv, // Sign it
				Topics: topics,
})
if err != nil {
	fmt.Println("could not create whisper envelope:", err)
}

fmt.Println("envelope:")
fmt.Println(envelope)

shh.Send(envelope)

filterTopics := whisperv2.NewFilterTopicsFromStringsFlat("my","message")

shh.Watch(whisperv2.Filter{
        Topics: filterTopics,
        Fn:     func(msg *whisperv2.Message) {
								fmt.Println("recived message")
                fmt.Println(msg)
        },
})

	select {}
}
