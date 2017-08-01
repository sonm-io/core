FROM golang:1.8.3

WORKDIR /go/src/github.com/sonm-io/core
COPY . .

RUN make build_miner && cp ./sonmminer /sonmminer

ENTRYPOINT ["/sonmminer"]
