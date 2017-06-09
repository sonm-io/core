FROM golang:1.8.3

WORKDIR /go/src/github.com/sonm-io/insonmnia
COPY . .

RUN make install
