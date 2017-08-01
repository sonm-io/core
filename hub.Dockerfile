FROM golang:1.8.3

# Hub listen for TCP from Miners
EXPOSE 10002
# Hub gRPC API
EXPOSE 10001

WORKDIR /go/src/github.com/sonm-io/core
COPY . .

RUN make build_hub && cp ./sonmhub /sonmhub

ENTRYPOINT ["/sonmhub"]