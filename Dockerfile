FROM golang:1.8.3

# Hub listen for TCP from Miners
EXPOSE 10002
# Hub gRPC API
EXPOSE 10001

WORKDIR /go/src/github.com/sonm-io/insonmnia
COPY . .

RUN make install
