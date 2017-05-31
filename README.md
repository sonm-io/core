
# Fusrodah

```DRAGONBORN!```

SONM messages system. Based on Ethereum Whisper.

## Fusrodah

Fusrodah is universal clinet which can both broadcast and listen messages with specified topics.
Can be used as prototype.

Programm is creates itsown p2p server (simple node) and launch sub-protocol whisper(v2)

### Fusrodah build from source

1. ``` golang 1.7 required ```
2. Install hacked library first
3. Clone this repo
4. inside clonned directory run ```go run fusrodah.go```




## Hacked library

NOTE - you should understand that program used **not original** go-ethereum library but this hacked version.
https://github.com/sonm-io/go-ethereum

### How install hacked library for working with project?

1. Install 'hacked' go-ethereum by ```go get github.com/sonm-io/go-ethereum ```
3. ???
4. PROFIT!!!

### Description of the system cycle

1) Function 'start' starts the Whisper server. This step requires a private key.
Then the function creates a new instance of a Whisper protocol entity.
NOTE - using Whisper v.2 (not v5).A configuration is set to a running p2p server.
Configuration values can't be modified after launch.
See the p2p package in go-ethereum (server.go) for more info.
2) The function then defines the p2p server and binds it to the configuration.
The configuration can also be stored in file. The server starts and listens for errors.
The Whisper protocol then launches on the running server (the process is usually automated).
3) Function "getFilterTopics" creates new filters for new topics.
4) Function "createMessage" creates a message entity.
Messages are represented by an end-user data packet, which is then transmitted via the Whisper protocol.
These are wrapped into Envelopes that need not be understood by intermediate nodes, just forwarded.
NewMessage creates and initializes a non-signed, non-encrypted Whisper message.
5) Function "createEnvelope" wraps the message into an envelope. This creates a bundle with the message inside,
which is then transferred via the network. PoW (Proof Of Work) controls how much time is spent on hashing the message,
inherently controlling its priority throughout the network (smaller hash, bigger priority).
6) Function “Send” starts the Whisper server, if it is not running yet, after which it wraps the source message,
sends it to *whisper2.Message and get possible topics for it.
7) Function "AddHandling" starts the Whisper server, if it is not running yet, and adds a watcher with any topics for it to handle.

In summation, we need a safe configuration using private keys provided by miners for encryption.
The testsFn function generates a standard private key, while the keystore.newKey generates an ethereum key
(a key with an ethereum address, etc.)
The main program logic is split along the two paths: the miner logic and the hub logic.
In the miner catalogue you can find the filtration functions available for the miner,
which can then be used to get a list of hubs, sorted via a variety of filters.


