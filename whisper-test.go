package main

/*

	This program use modified go-ethereum libraris, which can be finded in modified_library directory

	Author Sergey Ponomarev (@JackBekket on GitHub)


	Copiright 2017
*/

import (
	"fmt"
	//"log"
	"os"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/whisper/whisperv2"
	//"github.com/ethereum/go-ethereum/crypto/secp256k1"

	"github.com/ethereum/go-ethereum/crypto"


)

func main() {



	//This is generate standart private key..(just private ket, NOT ethereum key struct.)
	//For generating ethereum key struct (with ethereum address etc) - use keystore.newKey
	// Private key is secp256k1
	prv, _ :=crypto.GenerateKey()


	//Creates new instance of whisper protocol entity. NOTE - using whisper v.2 (not v5)
	shh := whisperv2.New()




//-----------P2P SERVER---------------------------------------------------------

/*

	This part is uses for configure and launch p2p server (p101/ethereum standart)
	and launch additional protocols above it.


*/

	//Configuration to running p2p server. Configuration values can't be modified after launch.
	//See p2p package in go-ethereum (server.go) for more info.
	cfg := p2p.Config{
		MaxPeers:   10,
	//	Identity:   p2p.NewSimpleClientIdentity("my-whisper-app", "1.0", "", string(pub)),
		PrivateKey: prv,
		ListenAddr: ":8000",

		//here we can define what additional protocols will be used *above* p2p server.
		Protocols: []p2p.Protocol{shh.Protocol},
	}

	//Definition of p2p server and binds to configuration. Configuration also could be stored in file.
	srv:= p2p.Server{
		Config: cfg,
	}



  //srv.Start();

	//Starting server and listen to errors.
	if err := srv.Start(); err != nil {
		fmt.Println("could not start server:", err)
	//	srv.Stop()
		os.Exit(1)
	}


	//Starting whisper protocol on running server.
	// NOTE whisper *should* be started automatically but it is not happening... possible BUG in go-ethereum.
	if err:=	shh.Start(srv); err != nil {
		fmt.Println("could not start whisper:", err)
	//	srv.Stop()
		os.Exit(1)
	}

//------------------------------------------------------------------------------





//---------MESSAGES LOGIC-------------------------------------------------------

/*

	This part is implement messages logic itself.


*/

//--------SEND MESSAGE----------------------------------------------------------
/*

	This part is uses to send messages throught whisper protocol

*/


// NOTE for single topic use NewTopicFromString
// NOTE whisperv2 is a package, shh - running whisper entity. Do not mess with that.
// NOTE topics logic can be finded in whisperv2/topic.go
// Topic represents a cryptographically secure, probabilistic partial
// classifications of a message, determined as the first (left) 4 bytes of the
// SHA3 hash of some arbitrary data given by the original author of the message.
topics := whisperv2.NewTopicsFromStrings("my", "message")

// This just print encoded topics. I use it for test, so this should be removed
// to production ver.
fmt.Println("topics")
fmt.Println(topics)



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
msg := whisperv2.NewMessage([]byte("hello world"))  // 1

// This print entity of new whisper message
// Just for test, should be removed in production
fmt.Println("msg:")
fmt.Println(msg)

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
envelope, err := msg.Wrap(1,whisperv2.Options{                // 2
				From:   prv, // Sign it
				Topics: topics,
})
if err != nil {
	fmt.Println("could not create whisper envelope:", err)
}

// Print envelope. just for test version, remove it to production.
fmt.Println("envelope:")
fmt.Println(envelope)

// Sending wrapped message.
shh.Send(envelope)
//------------------------------------------------------------------------------


//--------------Listen Message--------------------------------------------------
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
