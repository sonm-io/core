FROM golang:1.8.3

# Bootnode http info page
EXPOSE 8092

WORKDIR /go/src/github.com/sonm-io/core
COPY . .

RUN make build_bootnode && cp ./sonmbootnode /sonmbootnode

ENTRYPOINT ["/sonmbootnode"]
