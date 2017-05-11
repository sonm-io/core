# Hacked go ethereum

There is hacked go-ethereum library.

This package is modified and configured best way to use whisper protocol for
native go-lang applications.


## Changelog

### package p2p

 /```server.go```

 Added ```fmt.Println``` for each ```log``` to get logs info in main go-app.
 (Should be removed when we will have logs handler inside main app)

### set.v0
External dependency from gopkg.in is transferred inside
library root cause of broken original namespace and import config.

### whisper

/```peer.go```
fixed broken import of set.v0

/```whisper.go```

 fixed broken dependency (set.v0)
 **line 65,95,118,123 - fixed p2p.Protocol method inside Whisper struct to be able to use
 outside go-ethereum**
 added fmt.PrintLn for some logs.
 
