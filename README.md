
# Shadow Word

SONM messages system. Based on Ethereum Whisper.

## Fusrodah

Fusrodah is universal clinet which can both broadcast and listen messages with specified topics.
Can be used as prototype.

Programm is creates itsown p2p server (simple node) and launch sub-protocol whisper(v2)

### Fusrodah build from source

1. Install hacked library first
2. Clone this repo
3. inside clonned directory run ```go run fusrodah.go```
4.``` golang 1.7 required ```




## Hacked library

NOTE - you should understand that program used **not original** go-ethereum library but this hacked version.
You can get more in modified_library README.md

### How install hacked library for working with project?

1. Install standart go ethereum by ```go get github.com/ethereum/go-ethereum ```
2. Replace go-ethereum from directory modified_library from this repo to GOPATH/src/github.com/ethereum/
3. ???
4. PROFIT!!!
