.PHONY: clean bootnode fusrodah fusrodah_direct hub miner


default: bootnode fusrodah fusrodah_direct hub miner

check: default clean

bootnode:
	go build -tags nocgo -o bootnode_build ./bootnode/main.go

fusrodah:
	go build -tags nocgo -o fusrodah_build ./fusrodah/main.go

fusrodah_direct:
	go build -tags nocgo -o fusrodah_direct_build ./fusrodah_direct/main.go

hub:
	go build -tags nocgo -o hub_build ./hub/main.go

miner:
	go build -tags nocgo -o miner_build ./miner/main.go

clean:
	rm -f *_build