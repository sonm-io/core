# Buld stage
FROM frolvlad/alpine-glibc:latest

COPY ./sonmnode /usr/bin/sonmnode

RUN apk update && \
    apk add ca-certificates && \
    chmod +x /usr/bin/sonmnode

WORKDIR /etc/sonm

ENTRYPOINT /usr/bin/sonmnode --config=/etc/sonm/node-default.yaml

