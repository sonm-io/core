FROM golang:1.8.3

# Hub listen
EXPOSE 10002
# Miner listen
EXPOSE 10001

WORKDIR /go/src/github.com/sonm-io/insonmnia
COPY . .

RUN make install
